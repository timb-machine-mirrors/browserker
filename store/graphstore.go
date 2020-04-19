package store

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
)

type Graph struct {
	Store    *cayley.Handle
	filepath string
	dbType   string
}

func NewGraph(dbType, filepath string) *Graph {
	return &Graph{dbType: dbType, filepath: filepath}
}

func (g *Graph) Init() error {
	var err error

	// Initialize the database
	err = graph.InitQuadStore(g.dbType, g.filepath, nil)
	if err != nil {
		return err
	}

	// Open and use the database
	g.Store, err = cayley.NewGraph(g.dbType, g.filepath, nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *Graph) Load(path string) error {
	return g.Init()
}

func (g *Graph) Close() error {
	if g.Store != nil {
		return g.Store.Close()
	}
	return nil
}
