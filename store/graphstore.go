package store

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"

	// required by cayley for registering stores of diff types
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
)

// InitGraph or open existing
func InitGraph(dbType, filepath string) (*cayley.Handle, error) {
	var err error

	// Initialize the database
	err = graph.InitQuadStore(dbType, filepath, nil)
	if err != nil && err != graph.ErrDatabaseExists {
		return nil, err
	}

	// Open and use the database
	store, err := cayley.NewGraph(dbType, filepath, nil)
	if err != nil {
		return nil, err
	}
	return store, nil
}

// Close the db handle
func Close(handle *cayley.Handle) error {
	return handle.Close()
}
