package dit

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"fmt"
	"os"
	"sync"

	"github.com/cuhsat/go-secretsdump/pkg/system"

	"github.com/cuhsat/go-secretsdump/pkg/ese"
)

// New Creates a new dit dumper
func New(system, ntds string) (Reader, error) {
	r := Reader{
		isRemote:           false,
		history:            false,
		noLMHash:           true,
		remoteOps:          "",
		useVSSMethod:       false,
		justNTLM:           false,
		pwdLastSet:         false,
		resumeSession:      "",
		outputFileName:     "",
		justUser:           "",
		printUserStatus:    false,
		systemHiveLocation: system,
		ntdsFileLocation:   ntds,
		userData:           make(chan Info, 500),
	}

	var err error
	r.db, err = ese.New(ntds)
	if err != nil {
		return r, err
	}
	r.cursor, err = r.db.OpenTable("datatable")
	if err != nil {
		return r, err
	}
	//go r.dump() //start dumping the file immediately output will be put into the output channel as it comes

	return r, err
}

type Reader struct {
	ntdsFile  *os.File
	bootKey   []byte
	isRemote  bool
	history   bool
	noLMHash  bool
	remoteOps string

	useVSSMethod       bool
	justNTLM           bool
	pwdLastSet         bool
	resumeSession      string
	outputFileName     string
	systemHiveLocation string
	ntdsFileLocation   string

	justUser        string
	printUserStatus bool

	perSecretCallback bool // nil
	secret            bool //nil

	resumeSessionMgr bool // nil

	db       *ese.Esedb
	cursor   *ese.Cursor
	pek      [][]byte
	tmpUsers []ese.Esent_record

	//output chans
	userData    chan Info
	decryptWork chan ese.Esent_record
	cryptwg     *sync.WaitGroup

	//settings Settings
}

// Chan returns a reference to the objects output channel for read only operations
func (d *Reader) Chan() <-chan Info {
	return d.userData
}

func (d *Reader) Dump() error {
	var err error

	//if local (always local for now)
	if d.systemHiveLocation != "" {
		ls := system.New(d.systemHiveLocation)
		d.bootKey, err = ls.BootKey()

		if err != nil {
			return err
		}

		if d.ntdsFileLocation != "" {
			d.noLMHash = ls.HasNoLMHashPolicy()
		}
	} else {
		return fmt.Errorf("system hive empty")
	}

	_, _ = d.getPek()
	if len(d.pek) < 1 {
		return fmt.Errorf("NO PEK FOUND THIS IS VERY BAD")
	}

	for {
		//read each record from the db
		record, err := d.db.GetNextRow(d.cursor)
		if err != nil {
			if err.Error() == "ignore" {
				break //we will get an 'ignore' error when there are no more records
			}
			fmt.Println("Couldn't get row due to error: ", err.Error())
			continue
		}

		//check for the right kind of record
		v, ook := record.GetLongVal(nsAMAccountType)
		if ook {
			if _, ok := accTypes[v]; ok {
				dh, err := d.DecryptRecord(record)
				if err != nil {
					fmt.Println("Coudln't decrypt record:", err.Error())
					continue
				}
				d.userData <- dh

			}
		}
	}
	close(d.userData)
	return nil
}

func (d *Reader) getPek() ([][]byte, error) {
	var pekList []byte
	for {
		record, err := d.db.GetNextRow(d.cursor)
		if err != nil && err.Error() != "ignore" {
			return nil, err
		}
		if err != nil && err.Error() == "ignore" {
			break //lol fml
		}

		if v, ok := record.GetBytVal(npekList); ok && len(v) > 0 {
			//if v, ok := record.Column[pekList"]]; ok && len(v.BytVal) > 0 {
			pekList = v
			break
		}

		if r := record.GetNilRecord(nsAMAccountType); r != nil {
			//users found?
			d.tmpUsers = append(d.tmpUsers, record)
		}
	}
	if len(pekList) > 0 { //not an empty pekkyboi

		encryptedPekList, err := NewPeklistEnc(pekList)
		if err != nil {
			//should probably hard fail here
			return nil, err
		}
		if bytes.Compare(encryptedPekList.Header[:4], []byte{2, 0, 0, 0}) == 0 {
			//up to Windows 2012 r2 something
			md := md5.New()
			md.Write(d.bootKey)
			for i := 0; i < 1000; i++ {
				md.Write(encryptedPekList.KeyMaterial[:])
			}
			tmpKey := md.Sum([]byte{})
			rc, err := rc4.NewCipher(tmpKey)
			if err != nil {
				return nil, err
			}
			dst := make([]byte, len(encryptedPekList.EncryptedPek))
			rc.XORKeyStream(dst, encryptedPekList.EncryptedPek)
			decryptedPekList := NewPeklistPlain(dst)
			pekLen := 20 //len of the pek_key structure
			for i := 0; i < len(decryptedPekList.DecryptedPek)/pekLen; i++ {
				cursor := i * pekLen
				//fmt.Println("PEK found and decrypted:", hex.EncodeToString(pek.Key[:]))
				d.pek = append(d.pek, NewPekKey(decryptedPekList.DecryptedPek[cursor:cursor+pekLen]))
			}

		} else if bytes.Compare(encryptedPekList.Header[:4], []byte("\x03\x00\x00\x00")) == 0 {
			// something 2016 TP4
			/*
				# Windows 2016 TP4 header starts this way
				# Encrypted PEK Key seems to be different, but actually similar to decrypting LSA Secrets.
				# using AES:
				# Key: the bootKey
				# CipherText: PEKLIST_ENC['EncryptedPek']
				# IV: PEKLIST_ENC['KeyMaterial']
			*/
			ePek, err := DecryptAES(d.bootKey, encryptedPekList.EncryptedPek, encryptedPekList.KeyMaterial[:])
			if err != nil {
				return nil, err
			}
			decryptedPekList := NewPeklistPlain(ePek)
			d.pek = append(d.pek, decryptedPekList.DecryptedPek[4:20])
		}
	}
	return d.pek, nil
}
