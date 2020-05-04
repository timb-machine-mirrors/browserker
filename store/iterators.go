package store

import (
	badger "github.com/dgraph-io/badger/v2"
	"gitlab.com/browserker/browserker"
)

func StateIterator(txn *badger.Txn, byState browserker.NavState, limit int64) ([][]byte, error) {
	states := make([][]byte, limit)
	idx := int64(0)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: []byte("state:")})
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		if idx == limit {
			break
		}

		item := it.Item()
		val, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		state, err := DecodeState(val)
		if err != nil {
			return nil, err
		}
		
		if state == byState {
			states[idx] = GetID(item.KeyCopy(nil))
			idx++
		}
	}
	return states, nil
}