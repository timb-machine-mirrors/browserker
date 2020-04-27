package store

import "github.com/cayleygraph/cayley"

type CrawlGraph struct {
	Store    *cayley.Handle
	filepath string
	dbType   string
}

func NewCrawlGraph(dbType, filepath string) *CrawlGraph {
	return &CrawlGraph{dbType: dbType, filepath: filepath}
}

func (g *CrawlGraph) Init() error {
	var err error

	g.Store, err = InitGraph(g.dbType, g.filepath)
	return err
}
