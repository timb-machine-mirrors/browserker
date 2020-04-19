package store_test

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/simpscan/store"
)

func TestInit(t *testing.T) {
	dir, err := ioutil.TempDir("testdata/", "tests")
	if err != nil {
		t.Fatalf("error opening testdir: %s\n", err)
	}
	defer os.RemoveAll(dir)
	s := store.NewGraph("bolt", dir)
	s.Close()
}
