package ese

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

//props to agsolino for doing the original impacket version of this. The file format is clearly a mindfuck, and it would not have been easy.

// todo: update to handle massive files better (so we don't saturate memory too bad)
type fileInMem struct {
	//data  []byte
	pages []*page
}

type Ese struct {
	//options?
	reader       io.Reader
	pageSize     uint32
	db           *fileInMem
	dbHeader     dbHeader
	totalPages   uint32
	tables       map[string]*table
	currentTable string
	isRemote     bool
}

var stringCodePages = map[uint32]string{
	1200:  "utf-16le",
	20127: "ascii",
	1252:  "cp1252",
} //standin for const lookup/enum thing

func New(r io.Reader, n int) (*Ese, error) {
	db := &Ese{
		reader:   r,
		pageSize: pageSize,
		tables:   make(map[string]*table),
		isRemote: false,
	}

	err := db.parse(r, n)

	return db, err
}

// OpenTable opens a table, and returns a cursor pointing to the current parsing state
func (e *Ese) OpenTable(s string) (*Cursor, error) {
	r := Cursor{} //this feels like it can be optimized

	//if the table actually exists
	if v, ok := e.tables[s]; ok {
		entry := v.TableEntry
		//set the header
		ddHeader := dataDefinitionHeader{}
		buffer := bytes.NewBuffer(entry.EntryData)
		err := binary.Read(buffer, binary.LittleEndian, &ddHeader)
		if err != nil {
			return nil, err
		}

		//initialize the catalog entry
		catEnt, err := dataDefinitionEntry{}.Init(entry.EntryData[4:])
		if err != nil {
			return nil, err
		}

		//determine the page number to retreive
		pageNum := catEnt.Other.FatherDataPageNumber
		var page *page
		var done = false
		for !done {
			page = e.getPage(pageNum)
			if page.record.FirstAvailablePageTag <= 1 {
				//no records
				break
			}
			for i := uint16(1); i < page.record.FirstAvailablePageTag; //goland:noinspection GoUnreachableCode
			i++ {
				if page.record.PageFlags&FlagsLeaf == 0 {
					flags, data, err := page.getTag(int(i))
					if err != nil {
						//TODO: decide if we want to error, or die
						fmt.Println(err)
					}
					branchEntry := branchEntry{}.Init(flags, data)
					pageNum = branchEntry.ChildPageNumber
					break
				} else {
					done = true
					break
				}
			}
		}
		cursor := Cursor{
			TableData:            e.tables[s],
			FatherDataPageNumber: catEnt.Other.FatherDataPageNumber,
			CurrentPageData:      page,
			CurrentTag:           0,
		}
		return &cursor, nil
	}

	return &r, nil
}

func (e *Ese) parse(r io.Reader, n int) (err error) {
	//the first page is the dbheader
	err = e.loadPages(r, n)
	if err != nil {
		return
	}
	// this was a gross way of working out how many pages the file has...
	//this is where everything actually gets parsed out
	_ = e.parseCatalog(CatalogPageNumber) //4  ?

	return
}

func (e *Ese) parseCatalog(pagenum uint32) error {
	//parse all pages starting at pagenum, and add to the in-memory table

	//get the page
	page := e.getPage(pagenum)

	//parse the page
	_ = e.parsePage(page)

	//Iterate over each tag in the branch
	for i := 1; i < int(page.record.FirstAvailablePageTag); i++ {
		//get the tag flags, and data
		flags, data, err := page.getTag(i)
		if err != nil {
			return err
		}
		//if we are looking at a branch page
		if //goland:noinspection GoMaybeNil
		page.record.PageFlags&FlagsLeaf == 0 {
			//create the branch entry from the flags and data retreived
			branchEntry := branchEntry{}.Init(flags, data)
			//walk along the branch, and parse any referenced pages
			_ = e.parseCatalog(branchEntry.ChildPageNumber)
		}
	}
	return nil
}
func (e *Ese) parsePage(page *page) error {
	//baseOffset := page.record.Len // useless line?
	if page.record.PageFlags&FlagsLeaf == 0 || //not a leaf, don't care
		page.record.PageFlags&FlagsLeaf > 0 && (page.record.PageFlags&FlagsSpaceTree > 0 ||
			page.record.PageFlags&FlagsIndex > 0 || page.record.PageFlags&FlagsLongValue > 0) {
		return nil
	}

	//must be table entry
	for tagnum := 1; tagnum < int(page.record.FirstAvailablePageTag); tagnum++ {
		flags, data, err := page.getTag(tagnum)
		if err != nil {
			return err
		}
		leafEntry := leafEntry{}.Init(flags, data)
		_ = e.addLeaf(leafEntry)
	}
	return nil
}

func (e *Ese) GetNextRow(c *Cursor) (Record, error) {
	c.CurrentTag++
	// increment cursor pointer to look for 'next' tag

	//getnexttag starts here
	page := c.CurrentPageData

	if page == nil || c.CurrentTag >= uint32(page.record.FirstAvailablePageTag) ||
		//err = errors.New("ignore") //nil
		(page.record.PageFlags&FlagsLeaf == 0 || //not a leaf, don't care
			page.record.PageFlags&FlagsLeaf > 0 && (page.record.PageFlags&FlagsSpaceTree > 0 ||
				page.record.PageFlags&FlagsIndex > 0 || page.record.PageFlags&FlagsLongValue > 0)) {

		if page == nil || page.record.NextPageNumber == 0 { //no more pages :(
			return Record{}, errors.New("ignore")
		}

		c.CurrentPageData = e.getPage(page.record.NextPageNumber)
		c.CurrentTag = 0
		return e.GetNextRow(c) //lol recursion
	}

	flags, data, err := page.getTag(int(c.CurrentTag))
	if err != nil {
		return Record{}, err
	}
	tag := leafEntry{}.Init(flags, data)
	return e.tagToRecord(c, tag.EntryData)
}

func (e *Ese) addLeaf(l leafEntry) error {
	ddHeader := dataDefinitionHeader{}
	buffer := bytes.NewBuffer(l.EntryData)
	err := binary.Read(buffer, binary.LittleEndian, &ddHeader)
	if err != nil {
		return err
	}

	ce, err := dataDefinitionEntry{}.Init(l.EntryData[4:])
	if err != nil {
		//can't parse the entry good, ignore it lol
		return err
	}

	itemName, err := e.parseItemName(l)
	if err != nil {
		return err
	}
	//create table
	if ce.Fixed.Type == CatalogTypeTable {
		//t := newTable(string(itemName))
		///*
		t := table{}
		t.TableEntry = l
		t.Columns = &categoryEntries{} // make(map[string]cat_entry)
		//t.Indexes = &OrderedMap_esent_leaf_entry{values: make(map[string]esent_leaf_entry)}    //make(map[string]esent_leaf_entry)
		//t.Longvalues = &OrderedMap_esent_leaf_entry{values: make(map[string]esent_leaf_entry)} //make(map[string]esent_leaf_entry)
		//*/
		//longvals
		e.tables[string(itemName)] = &t
		e.currentTable = string(itemName)
	} else if ce.Fixed.Type == CatalogTypeColumn {
		col := categoryEntry{

			leafEntry: leafEntry{
				CommonPageKeySize: l.CommonPageKeySize,

				LocalPageKeySize: l.LocalPageKeySize,
				LocalPageKey:     l.LocalPageKey,
				EntryData:        l.EntryData,
			},
			Header: ddHeader,
			Record: ce,
			Key:    string(itemName),
		}
		//e.tables[e.currentTable].AddColumn(string(itemName))
		e.tables[e.currentTable].Columns.Add(col)

	} else if ce.Fixed.Type == CatalogTypeIndex {

		//if e.tables[e.currentTable].Columns == nil {
		return nil
		//}
		//e.tables[e.currentTable].Indexes.Add(string(itemName), l)

	} else if ce.Fixed.Type == CatalogTypeLongValue {
		//
	} else {
		return fmt.Errorf("reached code it shuldn't")
	}
	return nil
}

func (e *Ese) parseItemName(l leafEntry) ([]byte, error) {
	ddHeader := dataDefinitionHeader{}
	buffer := bytes.NewBuffer(l.EntryData)
	err := binary.Read(buffer, binary.LittleEndian, &ddHeader)
	if err != nil {
		return nil, err
	}
	entries := uint8(0)
	if ddHeader.LastVariableDataType > 127 {
		entries = ddHeader.LastVariableDataType - 127
	} else {
		entries = ddHeader.LastVariableDataType
	}
	entryLen := binary.LittleEndian.Uint16(l.EntryData[ddHeader.VariableSizeOffset:][:2])
	entryName := l.EntryData[ddHeader.VariableSizeOffset:][2*entries:][:entryLen]
	return entryName, err
}

func (e *Ese) getMainHeader(data []byte) (dbHeader, error) {
	dbhd := dbHeader{}
	buffer := bytes.NewBuffer(data)
	err := binary.Read(buffer, binary.LittleEndian, &dbhd)
	return dbhd, err
}

func (e *Ese) loadPages(r io.Reader, n int) error {
	var err error
	e.db = &fileInMem{}
	fr := bufio.NewReader(r)
	hdr := make([]byte, e.pageSize)
	_, _ = fr.Read(hdr)
	e.dbHeader, err = e.getMainHeader(hdr)
	if err != nil {
		return err
	}
	e.pageSize = e.dbHeader.PageSize

	pages := n / int(e.pageSize)
	e.db.pages = make([]*page, pages)
	e.totalPages = uint32(pages - 2) //unsure why -2 at this stage, I assume first page is header and last page is tail?

	for i := uint32(1); i < e.totalPages; i++ {
		start := i * e.pageSize
		if int(start) > n {
			return nil
		}
		end := start + e.pageSize
		//r := make([]byte, count)
		if int(end) > n {
			end = uint32(n)
		}
		e.db.pages[i] = &page{data: make([]byte, e.pageSize)}
		r := e.db.pages[i]
		_, _ = fr.Read(r.data)
		r.dbHeader = e.dbHeader
		if r.data != nil {
			_ = r.getHeader()
		}
		r.cached = true
	}
	return nil
}

// retreives a page of data from the file?
func (e *Ese) getPage(pageNum uint32) *page {
	//check cache
	r := e.db.pages[pageNum+1]
	if r != nil {
		e.db.pages[pageNum+1] = nil
		return r
	}
	return nil //&esent_page{}
}
