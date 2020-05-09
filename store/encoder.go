package store

import (
	"bytes"
	"reflect"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/vmihailenco/msgpack/v4"
	"gitlab.com/browserker/browserk"
)

type PredicateField struct {
	key  []byte
	name string
}

// MakeKey of a predicate and id
func MakeKey(id []byte, predicate string) []byte {
	key := []byte(predicate)
	key = append(key, byte(':'))
	key = append(key, id...)
	return key
}

// GetID of key from a pred:key
func GetID(key []byte) []byte {
	split := bytes.SplitN(key, []byte(":"), 2)
	if len(split) == 1 {
		return []byte{}
	}
	return split[1]
}

// GetPredicate from pred:key
func GetPredicate(key []byte) []byte {
	split := bytes.SplitN(key, []byte(":"), 2)
	return split[0]
}

// Encode a struct reflect.Value denoted by index into a msgpack []byte slice
func Encode(val reflect.Value, index int) ([]byte, error) {
	return msgpack.Marshal(val.Field(index).Interface())
}

// EncodeState value
func EncodeState(state browserk.NavState) ([]byte, error) {
	return msgpack.Marshal(state)
}

// EncodeTime usually Now
func EncodeTime(t time.Time) ([]byte, error) {
	return msgpack.Marshal(t)
}

// DecodeNavigation takes a transaction and a nodeID and returns a navigation object or err
func DecodeNavigation(txn *badger.Txn, predicates []*NavGraphField, nodeID []byte) (*browserk.Navigation, error) {
	nav := &browserk.Navigation{}

	fields := make([]PredicateField, len(predicates))
	for i := 0; i < len(predicates); i++ {
		name := predicates[i].name
		key := MakeKey(nodeID, name)
		p := PredicateField{key: key, name: name}
		fields[i] = p
	}

	for _, pred := range fields {
		item, err := txn.Get(pred.key)
		if err != nil {
			return nil, err
		}
		if err := DecodeNavigationItem(item, nav, pred.name); err != nil {
			return nil, err
		}
	}

	return nav, nil
}

// DecodeNavigationItem of the predicate value into the navigation object.
// TODO autogenerate this
func DecodeNavigationItem(item *badger.Item, nav *browserk.Navigation, pred string) error {
	var err error
	switch pred {
	case "id":
		err = item.Value(func(val []byte) error {
			var b []byte
			err := msgpack.Unmarshal(val, &b)
			nav.ID = b
			return err
		})
	case "origin":
		err = item.Value(func(val []byte) error {
			var b []byte
			err := msgpack.Unmarshal(val, &b)
			nav.OriginID = b
			return err
		})
	case "trig_by":
		err = item.Value(func(val []byte) error {
			var v int16
			err := msgpack.Unmarshal(val, &v)
			nav.TriggeredBy = browserk.TriggeredBy(v)
			return err
		})
	case "state":
		err = item.Value(func(val []byte) error {
			var v int8
			err := msgpack.Unmarshal(val, &v)
			nav.State = browserk.NavState(v)
			return err
		})
	case "state_updated":
		err = item.Value(func(val []byte) error {
			var v time.Time
			err := msgpack.Unmarshal(val, &v)
			nav.StateUpdatedTime = v
			return err
		})
	case "dist":
		err = item.Value(func(val []byte) error {
			var v int
			err := msgpack.Unmarshal(val, &v)
			nav.Distance = v
			return err
		})
	case "scope":
		err = item.Value(func(val []byte) error {
			var v browserk.Scope
			err := msgpack.Unmarshal(val, &v)
			nav.Scope = v
			return err
		})
	case "action":
		err = item.Value(func(val []byte) error {
			v := &browserk.Action{}
			err := msgpack.Unmarshal(val, &v)
			nav.Action = v
			return err
		})
	default:
		panic("unknown predicate for navigation")
	}
	return err
}

func DecodeState(val []byte) (browserk.NavState, error) {
	var v int8
	err := msgpack.Unmarshal(val, &v)
	if err != nil {
		return browserk.NavInvalid, err
	}
	return browserk.NavState(v), nil
}

func DecodeID(val []byte) ([]byte, error) {
	var b []byte
	err := msgpack.Unmarshal(val, &b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
