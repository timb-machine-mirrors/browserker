package store

import "github.com/cayleygraph/cayley/quad"

type iterateFn func(value quad.Value)

type iterateCtx struct {
	Value     quad.Value
	IRI       quad.IRI
	index     int
	iteraters []iterateFn
}

func (c *iterateCtx) Next(value quad.Value) {
	for c.index < len(c.iteraters) {
		c.index++
		c.iteraters[c.index-1](value)
		return
	}
	c.index = 0
}

func (c *iterateCtx) Add(i ...iterateFn) {
	if c.iteraters == nil {
		c.iteraters = make([]iterateFn, 0)
	}
	c.iteraters = append(c.iteraters, i...)
}
