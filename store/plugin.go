package store

import (
	"os"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

// PluginStore saves plugin state and uniqueness
type PluginStore struct {
	Store    *badger.DB
	filepath string
}

// NewPluginStore for plugin storage
func NewPluginStore(filepath string) *PluginStore {
	return &PluginStore{filepath: filepath}
}

// Init the plugin state storage
func (s *PluginStore) Init() error {
	var err error

	if err = os.MkdirAll(s.filepath, 0677); err != nil {
		return err
	}

	s.Store, err = badger.Open(badger.DefaultOptions(s.filepath))

	if errors.Is(err, badger.ErrTruncateNeeded) {
		log.Warn().Msg("there was a failure re-opening database, trying to recover")
		opts := badger.DefaultOptions(s.filepath)
		opts.Truncate = true
		s.Store, err = badger.Open(opts)
	}

	if err != nil {
		return err
	}
	return nil
}

// IsUnique checks if a plugin event is unique for whatever uniqueFor is
func (s *PluginStore) IsUnique(evt *browserk.PluginEvent, uniqueFor browserk.UniqueFor) bool {
	return false
}

// AddEvent to the plugin store
func (s *PluginStore) AddEvent(evt *browserk.PluginEvent) {

}

// Close the plugin store
func (s *PluginStore) Close() error {
	return s.Store.Close()
}
