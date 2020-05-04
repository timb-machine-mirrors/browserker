package store

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserker"
)

func PathToNavIDs(txn *badger.Txn, predicates []*NavGraphField, nodeIDs [][]byte) ([][]*browserker.Navigation, error) {
	entries := make([][]*browserker.Navigation, len(nodeIDs))
	for idx, nodeID := range nodeIDs {
		entries[idx] = make([]*browserker.Navigation, 0)
		nav, err := DecodeNavigation(txn, predicates, nodeID)
		if err != nil {
			return nil, err
		}
		entries[idx] = append(entries[idx], nav) // add this one

		// walk origin
		log.Info().Msgf("!!!!!!!!WalkOrigin start for: %v", nodeID)
		if err := WalkOrigin(txn, predicates, entries[idx], nodeID); err != nil {
			return nil, err
		}
		log.Info().Msgf("!!!!!!!!WalkOrigin end for: %v----------", nodeID)
	}
	//TODO: Reverse entries
	return entries, nil
}

// WalkOrigin recursively walks back from a nodeID until we are at the root of the nav graph
func WalkOrigin(txn *badger.Txn, predicates []*NavGraphField, entries []*browserker.Navigation, nodeID []byte) error {
	log.Info().Msgf("WalkOrigin: %v", nodeID)
	if nodeID == nil || len(nodeID) == 0 {
		log.Info().Msgf("nodeID was nil")
		return nil
	}

	if len(entries) > 100 {
		return fmt.Errorf("max entries exceeded walking origin")
	}

	item, err := txn.Get(MakeKey(nodeID, "origin"))
	if err != nil {
		return err
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}

	id, err := DecodeID(val)
	if err != nil || len(id) == 0 {
		log.Info().Msgf("id was 0 for %#v", val)
		// origin id is empty signals we are done
		return nil
	}
	nav, err := DecodeNavigation(txn, predicates, id)
	if err != nil {
		return err
	}
	spew.Dump(nav)
	entries = append(entries, nav)
	return WalkOrigin(txn, predicates, entries, id)
}
