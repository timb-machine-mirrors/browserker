package store

import (
	"bytes"
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
	GraphStore          *badger.DB
	RequestStore        *badger.DB
	filepath            string
	navPredicates       []*NavGraphField
	navResultPredicates []*NavGraphField
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
	g.navResultPredicates = g.discoverPredicates(&browserk.NavigationResult{})
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

// AddNavigations entries into our graph and requests into request store
func (g *CrawlGraph) AddNavigations(navs []*browserk.Navigation) error {
	if navs == nil {
		return nil
	}

	return g.GraphStore.Update(func(txn *badger.Txn) error {
		for _, nav := range navs {

			if err := g.AddNavigation(nav); err != nil {
				log.Warn().Err(err).Msg("failed to add navigation")
				return nil
			}

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
		}
		return nil
	})
}

// NavExists check
func (g *CrawlGraph) NavExists(nav *browserk.Navigation) bool {
	var exist bool
	g.GraphStore.View(func(txn *badger.Txn) error {
		key := MakeKey(nav.ID, "state")
		value, _ := EncodeState(nav.State)

		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			log.Error().Err(err).Msg("failed to find node id state")
			return nil
		}

		retVal, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if bytes.Compare(value, retVal) == 0 {
			exist = true
		}
		return nil
	})
	return exist
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

// AddResult of a navigation. Iterate over all predicates and encode/store
// For the original navigation ID we want to store:
// r_nav_id:<nav id> = result.ID so we can GetNavigationResult(nav_id) to get
// the node ID for this result then look up <predicate>:resultID = ... values ...
func (g *CrawlGraph) AddResult(result *browserk.NavigationResult) error {

	return g.GraphStore.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(g.navResultPredicates); i++ {
			key := MakeKey(result.ID, g.navResultPredicates[i].name)
			rv := reflect.ValueOf(*result)
			bytez, err := Encode(rv, g.navResultPredicates[i].index)

			if g.navResultPredicates[i].name == "r_nav_id" {
				navKey := MakeKey(result.NavigationID, g.navResultPredicates[i].name)
				enc, _ := EncodeBytes(result.ID)
				// store this separately so we can it look it up (values are always encoded)
				txn.Set(navKey, enc)
			}

			if err != nil {
				log.Error().Err(err).Msg("failed to encode nav result")
				return err
			}
			// key = <id>:<predicate>, value = msgpack'd bytes
			txn.Set(key, bytez)
		}
		return nil
	})
}

// GetNavigationResult from the navigation id
func (g *CrawlGraph) GetNavigationResult(navID []byte) (*browserk.NavigationResult, error) {
	exist := &browserk.NavigationResult{}
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		var err error

		key := MakeKey(navID, "r_nav_id")
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			log.Error().Err(err).Msg("failed to find result navID")
			return nil
		}
		retVal, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		resultID, _ := DecodeID(retVal)
		exist, err = DecodeNavigationResult(txn, g.navResultPredicates, resultID)
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
			if nodeIDs == nil {
				log.Info().Msgf("No new nodeIDs")
				return nil
			}
			log.Info().Msgf("Found new nodeIDs for nav, getting paths: %#v", nodeIDs)
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

			if nodeIDs == nil {
				log.Info().Msgf("No new nodeIDs")
				return nil
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
