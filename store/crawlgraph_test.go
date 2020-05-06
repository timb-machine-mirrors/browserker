package store_test

import (
	"os"
	"testing"
	"time"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/store"
)

func TestCrawlGraph(t *testing.T) {
	os.RemoveAll("testdata/crawl")
	g := store.NewCrawlGraph("testdata/crawl")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	nav := testMakeNavi([]byte{0, 1, 2})
	nav.OriginID = []byte{}

	if err := g.AddNavigation(nav); err != nil {
		t.Fatalf("error adding: %s\n", err)
	}

	result, err := g.GetNavigation(nav.ID)
	if err != nil {
		t.Fatalf("error reading back navigation: %s\n", err)
	}

	if result.Action == nil {
		t.Fatalf("action was nil")
	}

	if string(nav.ID) != string(result.ID) {
		t.Fatalf("%v != %v\n", nav.ID, result.ID)
	}

	if nav.Action.Type != result.Action.Type {
		t.Fatalf("%v != %v\n", nav.Action.Type, result.Action.Type)
	}
	_ = g.Find(nil, browserk.NavUnvisited, browserk.NavInProcess, 5)
}

func TestCrawlAddMultiple(t *testing.T) {
	path := "testdata/multi/crawl"
	os.RemoveAll(path)

	g := store.NewCrawlGraph(path)
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	for i := 1; i < 11; i++ {
		nav := testMakeNavi([]byte{0, byte(i), 2})
		nav.OriginID = []byte{0, byte(i - 1), 2}
		nav.Distance = i - 1

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}

		if err := g.AddNavigation(nav); err != nil {
			t.Fatalf("error adding: %s\n", err)
		}
	}

	limit := 5
	entries := g.Find(nil, browserk.NavUnvisited, browserk.NavUnvisited, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	for i := 0; i < len(entries); i++ {
		t.Logf("%d %d", i, len(entries[i]))
		if len(entries[i]) != i+1 {
			t.Fatalf("entry should have len %d got %d\n", i+1, len(entries[i]))
		}

		last := entries[i][len(entries[i])-1]
		if last.Distance != i {
			t.Fatalf("last entry distance should be %d got %d\n", i, last.Distance)
		}
	}

	entries = g.Find(nil, browserk.NavUnvisited, browserk.NavInProcess, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	for i := 0; i < len(entries); i++ {
		t.Logf("%d %d", i, len(entries[i]))
		if entries[i][0].State != browserk.NavInProcess {
			t.Fatalf("expected in process state was %v\n", entries[i][0].State)
		}
	}
}

func testMakeNavi(id []byte) *browserk.Navigation {
	return &browserk.Navigation{
		ID:               id,
		StateUpdatedTime: time.Now(),
		TriggeredBy:      1,
		State:            browserk.NavUnvisited,
		Action: &browserk.Action{
			Type:   browserk.ActLoadURL,
			Input:  nil,
			Result: nil,
		},
	}
}
