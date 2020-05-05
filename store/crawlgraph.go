package store

import (
	"context"
	"os"
	"reflect"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

type NavGraphField struct {
	index int
	name  string
}

type CrawlGraph struct {
	GraphStore    *badger.DB
	RequestStore  *badger.DB
	filepath      string
	navPredicates []*NavGraphField
}

// NewCrawlGraph creates a new crawl graph and request store
func NewCrawlGraph(filepath string) *CrawlGraph {
	return &CrawlGraph{filepath: filepath}
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

	g.navPredicates = g.discoverPredicates(&browserk.Navigation{})
	return nil
}

func (g *CrawlGraph) discoverPredicates(f interface{}) []*NavGraphField {
	predicates := make([]*NavGraphField, 0)
	rt := reflect.TypeOf(f).Elem()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		fname := f.Tag.Get("graph")
		if fname != "" {
			predicates = append(predicates, &NavGraphField{
				index: i,
				name:  fname,
			})
		}
	}
	return predicates
}

// AddNavigation entry into our graph and requests into request store
func (g *CrawlGraph) AddNavigation(nav *browserk.Navigation) error {

	return g.GraphStore.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(g.navPredicates); i++ {
			key := MakeKey(nav.ID, g.navPredicates[i].name)

			rv := reflect.ValueOf(*nav)
			bytez, err := Encode(rv, g.navPredicates[i].index)
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
func (g *CrawlGraph) GetNavigation(id []byte) (*browserk.Navigation, error) {
	exist := &browserk.Navigation{}
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		var err error

		exist, err = DecodeNavigation(txn, g.navPredicates, id)
		return err
	})
	return exist, err
}

// Find navigation entries by a state. iff byState == setState will we not update the
// state (and time stamp) returns a slice of a slice of all navigations on how to get
// to the final navigation state (TODO: Optimize with determining graph edges)
func (g *CrawlGraph) Find(ctx context.Context, byState, setState browserk.NavState, limit int64) [][]*browserk.Navigation {
	// make sure limit is sane
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}

	entries := make([][]*browserk.Navigation, 0)
	if byState == setState {
		err := g.GraphStore.View(func(txn *badger.Txn) error {
			nodeIDs, err := StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}
			entries, err = PathToNavIDs(txn, g.navPredicates, nodeIDs)
			return err
		})

		if err != nil {
			log.Fatal().Err(err).Msg("failed to get path to navs")
		}
	} else {
		err := g.GraphStore.Update(func(txn *badger.Txn) error {
			nodeIDs, err := StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}

			err = UpdateState(txn, setState, nodeIDs)
			if err != nil {
				return err
			}
			entries, err = PathToNavIDs(txn, g.navPredicates, nodeIDs)
			return errors.Wrap(err, "path to navs")
		})

		if err != nil {
			log.Fatal().Err(err).Msg("failed to get path to navs")
		}
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
