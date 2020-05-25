package store

import (
	"bytes"

	badger "github.com/dgraph-io/badger/v2"
	"gitlab.com/browserker/browserk"
)

func StateIterator(txn *badger.Txn, byState browserk.NavState, limit int64) ([][]byte, error) {
	states := make([][]byte, 0)
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
			key := item.KeyCopy(nil)
			states = append(states, GetID(key))
			idx++
		}
	}
	// no entries
	if len(states) == 0 || states[0] == nil {
		return nil, nil
	}
	return states, nil
}

func IfIterator(txn *badger.Txn, key, value []byte, limit int64) ([][]byte, error) {
	results := make([][]byte, 0)
	idx := int64(0)
	it := txn.NewIterator(badger.IteratorOptions{Prefix: key})
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		if idx == limit {
			break
		}

		item := it.Item()
		retVal, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		if bytes.Compare(value, retVal) == 0 {
			results = append(results, GetID(item.KeyCopy(nil)))
			idx++
		}
	}
	if len(results) == 0 || results[0] == nil {
		return nil, nil
	}
	return results, nil
}
