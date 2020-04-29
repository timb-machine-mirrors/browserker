package store

import "github.com/cayleygraph/cayley"

type AttackGraph struct {
	Store    *cayley.Handle
	filepath string
	dbType   string
}

func NewAttackGraph(dbType, filepath string) *AttackGraph {
	return &AttackGraph{dbType: dbType, filepath: filepath}
}

func (g *AttackGraph) Init() error {
	var err error

	g.Store, err = InitGraph(g.dbType, g.filepath)
	return err
}

func (g *AttackGraph) AddAttack() {

}

func (g *AttackGraph) Close() error {
	return Close(g.Store)
}
