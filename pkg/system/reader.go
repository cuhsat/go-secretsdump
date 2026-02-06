package system

import (
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/text/encoding/unicode"
	"www.velocidex.com/golang/regparser"
)

type Reader struct {
	path string
}

func New(path string) *Reader {
	return &Reader{path: path}
}

func (r *Reader) BootKey() (key []byte, err error) {
	f, err := os.Open(r.path)

	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	reg, err := regparser.NewRegistry(f)

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

	var sbox = [16]int{8, 5, 4, 2, 11, 9, 13, 3, 0, 6, 1, 12, 14, 10, 15, 7}

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
	f, err := os.Open(r.path)

	if err != nil {
		return false
	}

	defer func() { _ = f.Close() }()

	reg, err := regparser.NewRegistry(f)

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
