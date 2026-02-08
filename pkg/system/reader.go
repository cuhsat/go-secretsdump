package system

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"golang.org/x/text/encoding/unicode"
	"www.velocidex.com/golang/regparser"
)

var sbox = [16]int{8, 5, 4, 2, 11, 9, 13, 3, 0, 6, 1, 12, 14, 10, 15, 7}

type Reader struct {
	reader *bytes.Reader
}

func New(r *bytes.Reader) *Reader {
	return &Reader{reader: r}
}

func (r *Reader) BootKey() (key []byte, err error) {
	reg, err := regparser.NewRegistry(r.reader)

	if err != nil {
		return nil, err
	}

	var b []byte

	for _, part := range []string{
		"JD", "Skew1", "GBG", "Data",
	} {
		key := reg.OpenKey(fmt.Sprintf("\\%s\\Control\\Lsa\\%s", r.controlSet(reg), part))
		buf := make([]byte, key.ClassLength())

		_, err = reg.BaseBlock.HiveBin().Reader.ReadAt(buf, int64(key.Class()+4096+4))

		if err != nil {
			return nil, err
		}

		b = append(b, buf...)
	}

	tmp := string(b)

	if len(b) > 32 {
		ud := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

		tmp, _ = ud.String(string(b))
	}

	sub, err := hex.DecodeString(tmp)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(sub); i++ {
		key = append(key, sub[sbox[i]])
	}

	return key, nil
}

func (r *Reader) HasNoLMHashPolicy() bool {
	reg, err := regparser.NewRegistry(r.reader)

	if err != nil {
		return false
	}

	key := reg.OpenKey(fmt.Sprintf("\\%s\\Control\\Lsa\\NoLmHash", r.controlSet(reg)))

	return key != nil
}

func (r *Reader) controlSet(reg *regparser.Registry) string {
	s := "ControlSet001"

	if k := reg.OpenKey("\\Select"); k != nil {
		for _, v := range k.Values() {
			if v.ValueName() == "Current" {
				s = fmt.Sprintf("ControlSet%03d", v.ValueData().Uint64)
			}
		}
	}

	return s
}
