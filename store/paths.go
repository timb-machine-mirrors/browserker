package store

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

func PathToNavIDs(txn *badger.Txn, predicates []*NavGraphField, nodeIDs [][]byte) ([][]*browserk.Navigation, error) {
	entries := make([][]*browserk.Navigation, len(nodeIDs))

	for idx, nodeID := range nodeIDs {
		// TODO INVESTIGATE WHY NODEIDS COULD BE EMPTY
		if len(nodeID) == 0 {
			break
		}
		entries[idx] = make([]*browserk.Navigation, 0)
		nav, err := DecodeNavigation(txn, predicates, nodeID)
		if err != nil {
			return nil, err
		}
		entries[idx] = append(entries[idx], nav) // add this one

		// walk origin
		if err := WalkOrigin(txn, predicates, &entries[idx], nodeID); err != nil {
			return nil, err
		}
		log.Info().Msgf("Before Reverse: %d", len(entries[idx]))
		// reverse the entries so we can crawl start to finish
		for i := len(entries[idx])/2 - 1; i >= 0; i-- {
			opp := len(entries[idx]) - 1 - i
			entries[idx][i], entries[idx][opp] = entries[idx][opp], entries[idx][i]
		}
		log.Info().Msgf("After Reverse: %d", len(entries[idx]))

	}
	return entries, nil
}

// WalkOrigin recursively walks back from a nodeID until we are at the root of the nav graph
func WalkOrigin(txn *badger.Txn, predicates []*NavGraphField, entries *[]*browserk.Navigation, nodeID []byte) error {
	log.Info().Msgf("WalkOrigin: %v", nodeID)
	if nodeID == nil || len(nodeID) == 0 {
		log.Info().Msgf("nodeID was nil")
		return nil
	}

	if len(*entries) > 100 {
		return fmt.Errorf("max entries exceeded walking origin")
	}

	item, err := txn.Get(MakeKey(nodeID, "origin"))
	if err != nil {
		// first & only node perhaps?
		return nil
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}

	id, err := DecodeID(val)
	if err != nil || len(id) == 0 {
		// origin id is empty signals we are done / at root node
		return nil
	}
	nav, err := DecodeNavigation(txn, predicates, id)
	if err != nil {
		return err
	}

	*entries = append(*entries, nav)
	log.Info().Msgf("WalkOrigin: len %d", len(*entries))
	return WalkOrigin(txn, predicates, entries, id)
}
