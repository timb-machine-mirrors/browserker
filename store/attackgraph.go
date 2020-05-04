package store

type AttackGraph struct {
	filepath string
}

func NewAttackGraph(filepath string) *AttackGraph {
	return &AttackGraph{filepath: filepath}
}

func (g *AttackGraph) Init() error {
	var err error
	return err
}

func (g *AttackGraph) AddAttack() {

}

func (g *AttackGraph) Close() error {
	return nil
}
