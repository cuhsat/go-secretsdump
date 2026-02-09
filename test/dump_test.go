package test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cuhsat/go-secretsdump/pkg/ntds"
	"github.com/cuhsat/go-secretsdump/pkg/system"
)

var bootkey = []byte{0x13, 0xd2, 0x09, 0x76, 0xd6, 0x3e, 0xa5, 0xe8, 0x36, 0x03, 0x6e, 0xc8, 0xbc, 0x68, 0xd6, 0xeb}

func TestDump(t *testing.T) {
	txt, err := os.ReadFile("data/ntds.txt")

	if err != nil {
		t.Fatal(err)
	}

	reg, err := os.ReadFile("./data/system")

	if err != nil {
		t.Fatal(err)
	}

	dit, err := os.ReadFile("./data/ntds.dit")

	if err != nil {
		t.Fatal(err)
	}

	key, err := system.New(bytes.NewReader(reg)).BootKey()

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(bootkey, key) {
		t.Errorf("invalid bootkey: %+x", key)
	}

	m := make(map[string]any, len(txt))

	for _, s := range strings.Split(string(txt), "\n") {
		if len(s) > 0 {
			m[s] = struct{}{}
		}
	}

	ch, err := ntds.New(
		bytes.NewReader(reg),
		bytes.NewReader(dit),
		len(dit),
	)

	if err != nil {
		t.Fatal(err)
	}

	for c := range ch {
		if _, ok := m[c.String()]; !ok {
			t.Errorf("unexpected hash: %s", c)
		}

		delete(m, c.String())
	}

	if len(m) > 0 {
		t.Errorf("expected hashes: %+v", m)
	}
}
