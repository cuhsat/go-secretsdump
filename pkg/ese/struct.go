package ese

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

type categoryEntry struct {
	leafEntry
	Header dataDefinitionHeader //??
	Record dataDefinitionEntry  //??
	Key    string
	KeyInt int
}
type table struct {
	Name       string
	TableEntry leafEntry
	Columns    *categoryEntries
}

type pageHeader struct {
	CheckSum                     uint64
	ECCCheckSum                  uint32
	LastModificationTime         uint64
	PreviousPageNumber           uint32
	NextPageNumber               uint32
	FatherDataPage               uint32
	AvailableDataSize            uint16
	AvailableUncommittedDataSize uint16
	FirstAvailableDataOffset     uint16
	FirstAvailablePageTag        uint16
	PageFlags                    uint32
	ExtendedCheckSum1            uint64
	ExtendedCheckSum2            uint64
	ExtendedCheckSum3            uint64
	PageNumber                   uint64
	Unknown                      uint64

	Len uint16
}

type jetSig struct {
	Random       uint32
	CreationTime uint64
	NetBiosName  [16]byte
}

type dbHeader struct {
	//thank god there is no dynamic fields in this structure
	CheckSum                   uint32
	Signature                  [4]byte //"\xef\xcd\xab\x89'),
	Version                    uint32
	FileType                   uint32
	DBTime                     uint64
	DBSignature                jetSig //:',ESENT_JET_SIGNATURE),
	DBState                    uint32
	ConsistentPosition         uint64
	ConsistentTime             uint64
	AttachTime                 uint64
	AttachPosition             uint64
	DetachTime                 uint64
	DetachPosition             uint64
	LogSignature               jetSig //:',ESENT_JET_SIGNATURE),
	Unknown                    uint32
	PreviousBackup             [24]byte
	PreviousIncBackup          [24]byte
	CurrentFullBackup          [24]byte
	ShadowingDisables          uint32
	LastObjectID               uint32
	WindowsMajorVersion        uint32
	WindowsMinorVersion        uint32
	WindowsBuildNumber         uint32
	WindowsServicePackNumber   uint32
	FileFormatRevision         uint32
	PageSize                   uint32
	RepairCount                uint32
	RepairTime                 uint64
	Unknown2                   [28]byte
	ScrubTime                  uint64
	RequiredLog                uint64
	UpgradeExchangeFormat      uint32
	UpgradeFreePages           uint32
	UpgradeSpaceMapPages       uint32
	CurrentShadowBackup        [24]byte
	CreationFileFormatVersion  uint32
	CreationFileFormatRevision uint32
	Unknown3                   [16]byte
	OldRepairCount             uint32
	ECCCount                   uint32
	LastECCTime                uint64
	OldECCFixSuccessCount      uint32
	ECCFixErrorCount           uint32
	LastECCFixErrorTime        uint64
	OldECCFixErrorCount        uint32
	BadCheckSumErrorCount      uint32
	LastBadCheckSumTime        uint64
	OldCheckSumErrorCount      uint32
	CommittedLog               uint32
	PreviousShadowCopy         [24]byte
	PreviousDifferentialBackup [24]byte
	Unknown4                   [40]byte
	NLSMajorVersion            uint32
	NLSMinorVersion            uint32
	Unknown5                   [148]byte
	UnknownFlags               uint32
}

type branchEntry struct {
	CommonPageKeySize uint16

	LocalPageKeySize uint16
	LocalPageKey     []byte // ":"
	ChildPageNumber  uint32
}

//goland:noinspection DuplicatedCode
func (e branchEntry) Init(flags uint16, data []byte) branchEntry {
	r := branchEntry{}
	//zzzz
	//data := make([]byte, len(ldata))
	//copy(data, ldata)
	//take first 2 bytes of data if common flag is set
	curs := 0
	if flags&TagCommon > 0 {
		r.CommonPageKeySize = binary.LittleEndian.Uint16(data[:2])
		//data = data[2:]
		curs += 2
	}
	//fill the structure with remaining data
	//first element is the pagekeysize
	r.LocalPageKeySize = binary.LittleEndian.Uint16(data[curs : curs+2])
	curs += 2
	//then the pagekey (determined by the pagekeysize)
	r.LocalPageKey = data[curs : curs+int(r.LocalPageKeySize)]
	curs += int(r.LocalPageKeySize)
	//data = data[curs+r.LocalPageKeySize:]
	//then we have the childpagenumber (this should be the rest of the data??)
	r.ChildPageNumber = binary.LittleEndian.Uint32(data[curs:])

	return r
}

type leafEntry struct {
	CommonPageKeySize uint16

	LocalPageKeySize uint16
	//_LocalPageKey    string //nil
	LocalPageKey []byte // ":"
	EntryData    []byte // ":"
}

//goland:noinspection DuplicatedCode
func (e leafEntry) Init(flags uint16, inData []byte) leafEntry {
	r := leafEntry{}
	curs := 0
	//data := make([]byte, len(inData))

	//copy(data, inData)
	//take first 2 bytes of data if common flag is set
	if flags&TagCommon > 0 {
		r.CommonPageKeySize = binary.LittleEndian.Uint16(inData[:2])
		//data = data[2:]
		curs += 2
	}
	//fill the structure with remaining data
	//first element is the pagekeysize
	r.LocalPageKeySize = binary.LittleEndian.Uint16(inData[curs : curs+2])
	curs += 2
	//then the pagekey (determined by the pagekeysize)
	r.LocalPageKey = inData[curs : curs+int(r.LocalPageKeySize)]
	curs += int(r.LocalPageKeySize)
	//data = data[r.LocalPageKeySize:]
	//then we have the data (this should be the rest of the data??)
	r.EntryData = inData[curs:]
	return r
}

type dataDefinitionHeader struct {
	LastFixedSize        uint8
	LastVariableDataType uint8
	VariableSizeOffset   uint16
}

type dataDefinitionEntry struct {
	Fixed   catalogFixedDataDefinitionEntry
	Columns catalogColumnsDataDefinitionEntry
	Other   catalogOtherDataDefinitionEntry
	Table   catalogTableDataDefinitionEntry
	Index   catalogIndexDataDefinitionEntry
	LV      catalogLvDataDefinitionEntry
	Common  catalogCommonDataDefinitionEntry
}

func (e dataDefinitionEntry) Init(inData []byte) (dataDefinitionEntry, error) {
	curs := 0
	r := dataDefinitionEntry{}
	//fill in fixed
	buffer := bytes.NewBuffer(getAndMoveCursor(inData, &curs, 10))
	err := binary.Read(buffer, binary.LittleEndian, &r.Fixed)
	if err != nil {
		return r, err
	}

	//this is where it gets hairy :(
	if r.Fixed.Type == CatalogTypeColumn {
		//only one with no 'other' section
		//fill in column stuff
		buffer := bytes.NewBuffer(getAndMoveCursor(inData, &curs, 16))
		err := binary.Read(buffer, binary.LittleEndian, &r.Columns)
		if err != nil {
			return r, err
		}
	} else {

		//fill in 'other'
		r.Other.FatherDataPageNumber = binary.LittleEndian.Uint32(getAndMoveCursor(inData, &curs, 4))

		if r.Fixed.Type == CatalogTypeTable {
			//do 'table stuff'
			r.Table.SpaceUsage = binary.LittleEndian.Uint32(getAndMoveCursor(inData, &curs, 4))
		} else if r.Fixed.Type == CatalogTypeIndex {
			//index stuff
			buffer := bytes.NewBuffer(getAndMoveCursor(inData, &curs, 12))
			err := binary.Read(buffer, binary.LittleEndian, &r.Index)
			if err != nil {
				return r, err
			}
		} else if r.Fixed.Type == CatalogTypeLongValue {
			r.LV.SpaceUsage = binary.LittleEndian.Uint32(getAndMoveCursor(inData, &curs, 4))
		} else if r.Fixed.Type == CatalogTypeCallback {
			return r, fmt.Errorf("catalog type callback unexpected")
		} else {
			return dataDefinitionEntry{}, errors.New("unkown Type")
		}
	}
	//fill in common stuff
	r.Common.Trailing = inData[curs:]

	return r, nil
}

func getAndMoveCursor(data []byte, curs *int, size int) []byte {
	d := data[*curs : *curs+size]
	*curs += size
	return d
}

type catalogFixedDataDefinitionEntry struct {
	FatherDataPageID uint32
	Type             uint16
	Identifier       uint32
}
type catalogColumnsDataDefinitionEntry struct {
	ColumnType  uint32
	SpaceUsage  uint32
	ColumnFlags uint32
	CodePage    uint32
}
type catalogOtherDataDefinitionEntry struct {
	FatherDataPageNumber uint32
}
type catalogTableDataDefinitionEntry struct {
	SpaceUsage uint32
}
type catalogIndexDataDefinitionEntry struct {
	SpaceUsage uint32
	IndexFlags uint32
	Locale     uint32
}
type catalogLvDataDefinitionEntry struct {
	SpaceUsage uint32
}
type catalogCommonDataDefinitionEntry struct {
	Trailing []byte
}

type Cursor struct {
	CurrentTag           uint32
	FatherDataPageNumber uint32
	CurrentPageData      *page
	TableData            *table
}

type Record struct {
	column map[string]*RecordValue
}

func NewRecord(i int) Record {
	return Record{column: make(map[string]*RecordValue, i)}
}

// SetString sets the codepage of the specified column on the record, and marks the record as a 'string'
func (e *Record) SetString(column string, codePage uint32) error {
	if e.column[column] == nil {
		//this should probably be a proper error
		return nil
	}
	if _, ok := stringCodePages[codePage]; !ok { //known decoding type
		return errors.New("unknown codepage")
	}
	e.column[column].SetString(codePage)
	return nil
}

func (e *RecordValue) SetString(codePage uint32) {
	e.typ = Str
	e.codePage = codePage
}

func (e *Record) GetLongVal(column string) (int32, bool) {
	v, ok := e.column[column]
	if v != nil && ok {
		return v.Long(), ok
	}
	return 0, ok
}

func (e *RecordValue) Long() int32 {
	return int32(binary.LittleEndian.Uint32(e.val))
}

func (e *Record) GetBytVal(column string) ([]byte, bool) {
	v, ok := e.column[column]
	if ok {
		return v.Bytes(), ok
	}
	return nil, ok
}

func (e *RecordValue) Bytes() []byte {
	return e.val
}

func (e *Record) StrVal(column string) (string, error) {
	v, ok := e.column[column]
	if ok {
		return v.String()
	}
	return "", fmt.Errorf("no value found")
}

var d = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

func (e *RecordValue) String() (string, error) {
	if e.codePage == 20127 { //ascii
		//v easy
		return string(e.val), nil
	} else if e.codePage == 1200 { // Unicode oh boy
		// Unicode utf16le

		b, err := d.Bytes(e.val)
		return string(b), err
		//record.Column[column] = Esent_recordVal{Typ: "Str", StrVal: string(b)}
	} else if e.codePage == 1252 {
		d := charmap.Windows1252.NewDecoder()
		b, err := d.Bytes(e.val)
		return string(b), err
		//western... idk yet
	}
	return "", fmt.Errorf("unknown codepage=%v", e.codePage)
}

func (e *Record) GetRecord(column string) *RecordValue {
	if r, ok := e.column[column]; ok {
		return r
	}
	r := &RecordValue{}
	e.column[column] = r
	return r
}

func (e *Record) GetNilRecord(column string) *RecordValue {
	if v, ok := e.column[column]; !ok {
		//e.column[column] = &Esent_recordVal{}
		return nil
	} else {
		return v
	}
}

type RecordValue struct {
	tupVal [][]byte
	val    []byte
	//strVal   string
	codePage uint32
	typ      recordTyp
}

// Data types possible in an ese database
//
//goland:noinspection GoUnusedConst
const (
	Byt recordTyp = iota
	Tup
	Str
	Nil
	Bit
	UnsByt
	Short
	Long
	Curr
	IEEESingl
	IEEEDoub
	DateTim
	Bin
	Txt
	LongBin
	LongTxt
	SLV
	UnsLng
	LngLng
	Guid
	UnsShrt
	Max
)

type recordTyp int

func (e *RecordValue) UpdateBytVal(d []byte) *RecordValue {
	e.typ = Byt
	//fmt.Println("record", d)
	e.val = d
	return e
}

func (e *RecordValue) UnpackInline(c catalogColumnsDataDefinitionEntry) {
	//if cRecord.Columns.ColumnType == JET_coltypText || cRecord.Columns.ColumnType == JET_coltypLongText {
	//record.SetString(column, cRecord.Columns.CodePage)
	t := c.ColumnType
	if len(e.val) < 1 {
		return
	}
	switch t {
	case JetColtypnil:
		e.typ = Nil
	case JetColtypbit:
		e.typ = Bit
	case JetColtypunsignedbyte:
		e.typ = UnsByt
	case JetColtypshort:
		e.typ = Short
	case JetColtyplong:
		e.typ = Long
	case JetColtypcurrency:
		e.typ = Curr
	case JetColtypieeesingle:
		e.typ = IEEESingl
	case JetColtypieeedouble:
		e.typ = IEEEDoub
	case JetColtypdatetime:
		e.typ = DateTim
	case JetColtypbinary:
		e.typ = Bin
	case JetColtyptext:
		e.typ = Txt
		e.SetString(c.CodePage)
	case JetColtyplongbinary:
		e.typ = LongBin
	case JetColtyplongtext:
		e.typ = LongTxt
		e.SetString(c.CodePage)
	case JetColtypslv:
		e.typ = SLV
	case JetColtypunsignedlong:
		e.typ = UnsLng
	case JetColtyplonglong:
		e.typ = LngLng
	case JetColtypguid:
		e.typ = Guid
	case JetColtypunsignedshort:
		e.typ = UnsShrt
	case JetColtypmax:
		e.typ = Max
	}
}

type taggedItem struct {
	TaggedOffset uint16
	TagLen       uint16
	Flags        uint16
}

type taggedItems struct {
	M []*taggedItem
	O []uint16
}

func (t *taggedItems) Add(tag *taggedItem, k uint16) {
	//NOT THREAD SAFE
	t.O = append(t.O, k)
	t.M = append(t.M, tag)
	//t.M[k] = &tag
}

type categoryEntries struct {
	values []categoryEntry
}

func (o *categoryEntries) Add(value categoryEntry) {
	o.values = append(o.values, value)
}

type OrderedLeafEntry struct {
	values map[string]leafEntry
	keys   []string
}

func (o *OrderedLeafEntry) Add(key string, value leafEntry) {
	_, exists := o.values[key]
	if !exists {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}
