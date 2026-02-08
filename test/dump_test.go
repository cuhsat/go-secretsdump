package test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cuhsat/go-secretsdump/pkg/ntds"
)

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
