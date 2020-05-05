package store

import (
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"gitlab.com/browserker/browserk"
)

func UpdateState(txn *badger.Txn, newState browserk.NavState, nodeIDs [][]byte) error {
	stateBytes, err := EncodeState(newState)
	if err != nil {
		return err
	}

	timeBytes, err := EncodeTime(time.Now())
	if err != nil {
		return err
	}

	for _, nodeID := range nodeIDs {
		if err := txn.Set(MakeKey(nodeID, "state"), stateBytes); err != nil {
			return err
		}
		if err := txn.Set(MakeKey(nodeID, "state_updated"), timeBytes); err != nil {
			return err
		}
	}
	return nil
}
