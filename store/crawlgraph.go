package store

import (
	"context"
	"os"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserker"
)

type NavGraphField struct {
	index int
	name  string
}

type CrawlGraph struct {
	GraphStore   *badger.DB
	RequestStore *badger.DB
	filepath     string
	predicates   []*NavGraphField
}

// NewCrawlGraph creates a new crawl graph and request store
func NewCrawlGraph(filepath string) *CrawlGraph {
	return &CrawlGraph{filepath: filepath, predicates: make([]*NavGraphField, 0)}
}

// Init the crawl graph and request store
func (g *CrawlGraph) Init() error {
	var err error

	if err = os.MkdirAll(g.filepath+"/requests", 0677); err != nil {
		return err
	}

	g.RequestStore, err = badger.Open(badger.DefaultOptions(g.filepath + "/requests"))
	if err != nil {
		return err
	}

	g.GraphStore, err = badger.Open(badger.DefaultOptions(g.filepath + "/crawl"))
	if err != nil {
		return err
	}

	g.discoverPredicates(&browserker.Navigation{})
	return nil
}

func (g *CrawlGraph) discoverPredicates(nav *browserker.Navigation) {
	rt := reflect.TypeOf(*nav)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		fname := f.Tag.Get("graph")
		if fname != "" {
			g.predicates = append(g.predicates, &NavGraphField{
				index: i,
				name:  fname,
			})
		}
	}
}

// AddNavigation entry into our graph and requests into request store
func (g *CrawlGraph) AddNavigation(nav *browserker.Navigation) error {

	return g.GraphStore.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(g.predicates); i++ {
			key := MakeKey(nav.ID(), g.predicates[i].name)

			rv := reflect.ValueOf(*nav)
			bytez, err := Encode(rv, g.predicates[i].index)
			if err != nil {
				return err
			}
			// key = <id>:<predicate>, value = msgpack'd bytes
			txn.Set(key, bytez)
		}
		return nil
	})
}

// GetNavigation by the provided id value
func (g *CrawlGraph) GetNavigation(id []byte) (*browserker.Navigation, error) {
	exist := &browserker.Navigation{}
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		var err error

		exist, err = DecodeNavigation(txn, g.predicates, id)
		return err
	})
	return exist, err
}

// Find navigation entries by a state. iff byState == setState will we not update the
// state (and time stamp) returns a slice of a slice of all navigations on how to get
// to the final navigation state (TODO: Optimize with determining graph edges)
func (g *CrawlGraph) Find(ctx context.Context, byState, setState browserker.NavState, limit int64) [][]*browserker.Navigation {
	// make sure limit is sane
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}

	entries := make([][]*browserker.Navigation, 0)
	if byState == setState {
		nodeIDs := make([][]byte, 0)
		err := g.GraphStore.View(func(txn *badger.Txn) error {
			var err error

			nodeIDs, err = StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}
			entries, err = PathToNavIDs(txn, g.predicates, nodeIDs)
			return err
		})

		if err != nil {
			log.Fatal().Err(err).Msg("failed to get path to navs")
		}
		spew.Dump(entries)
	}
	return entries
}

// Close the graph and request stores
func (g *CrawlGraph) Close() error {
	err1 := g.RequestStore.Close()
	err2 := g.GraphStore.Close()

	if err2 != nil {
		// print request store unable to close but just return the graph error
		if err1 != nil {
			log.Error().Err(err1).Msg("failed to close request store")
		}
		return err2
	}

	return err1
}
