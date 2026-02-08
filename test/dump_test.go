package test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cuhsat/go-secretsdump/pkg/ntds"
)

func TestDump(t *testing.T) {
	//get valid output files
	s, e := os.ReadFile("data/ntds.txt")
	if e != nil {
		t.Error("Could not read from file")
	}
	corretkerb := make(map[string]bool, len(s))
	sa := strings.Split(string(s), "\n")
	for _, v := range sa {
		if v != "" {
			corretkerb[v] = true
		}
	}

	b1, _ := os.ReadFile("./data/system")
	b2, _ := os.ReadFile("./data/ntds.dit")

	ch, err := ntds.New(bytes.NewReader(b1), bytes.NewReader(b2), len(b2))
	if err != nil {
		t.Fatal(err)
	}
	for ok := range ch {
		//ensure it exists (don't find values that are not in impacket... yet)
		if _, found := corretkerb[ok.String()]; !found {
			t.Errorf("found unexpected value: %s", ok.String())
		}
		//check history too
		for _, h := range ok.GetHistory() {
			if _, found := corretkerb[h]; !found {
				t.Errorf("found unexpected value: %s", h)
			}
			delete(corretkerb, h)
		}
		//ensure we don't miss any that impacket finds
		delete(corretkerb, ok.String())
	}
	if len(corretkerb) > 0 {
		t.Errorf("Expected empty map. Unfound hashes: %+v", corretkerb)
	}
}
