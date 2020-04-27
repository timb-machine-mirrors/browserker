package store_test

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/browserker/store"
)

func TestInit(t *testing.T) {
	dir, err := ioutil.TempDir("testdata/", "tests")
	if err != nil {
		t.Fatalf("error opening testdir: %s\n", err)
	}
	defer os.RemoveAll(dir)
	s, err := store.InitGraph("bolt", dir)
	if err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	s.Close()
}
