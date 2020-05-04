package store

import (
	"context"
	"fmt"
	"os"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/schema"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/browserker/browserker"
)

type NavEdge struct {
	Prev quad.IRI
	Next quad.IRI
}

type CrawlGraph struct {
	GraphStore   *cayley.Handle
	RequestStore *badger.DB
	cfg          *schema.Config
	filepath     string
	dbType       string
}

// NewCrawlGraph creates a new crawl graph and request store
func NewCrawlGraph(dbType, filepath string) *CrawlGraph {
	return &CrawlGraph{dbType: dbType, filepath: filepath}
}

// Init the crawl graph and request store
func (g *CrawlGraph) Init() error {
	var err error

	if err = os.MkdirAll(g.filepath+"/requests", 0677); err != nil {
		return err
	}

	g.RequestStore, err = badger.Open(badger.DefaultOptions(g.filepath + "/requests"))
	if err != nil {
		return err
	}

	g.GraphStore, err = InitGraph(g.dbType, g.filepath)
	g.cfg = schema.NewConfig()
	g.cfg.GenerateID = func(_ interface{}) quad.Value {
		return quad.IRI(uuid.NewV1().String())
	}
	schema.RegisterType("nav", browserker.Navigation{})
	return err
}

// AddNavigation entry into our graph and requests into request store
func (g *CrawlGraph) AddNavigation(nav *browserker.Navigation) error {
	writer := graph.NewWriter(g.GraphStore)
	id, err := g.cfg.WriteAsQuads(writer, nav)
	writer.Close()
	log.Logger.Info().Msgf("Wrote %s", id)
	return err
}

// GetNavigation by the provided id value
func (g *CrawlGraph) GetNavigation(id string) (*browserker.Navigation, error) {
	exist := &browserker.Navigation{}
	err := g.cfg.LoadTo(nil, g.GraphStore, exist, quad.IRI(id))
	return exist, err
}

// Find navigation entries by a state. iff byState == setState will we not update the
// state (and time stamp)
func (g *CrawlGraph) Find(ctx context.Context, byState, setState browserker.NavState, limit int64) map[string]*browserker.Navigation {
	entries := make(map[string]*browserker.Navigation, 0)

	p := cayley.StartPath(g.GraphStore).Has(quad.IRI("state"), quad.Int(byState)).Order()
	if limit != 0 {
		p.Limit(limit)
	}

	log.Info().Msg("calling iterate")

	tx := cayley.NewTransaction()
	ci := &iterateCtx{}
	ci.Add(g.GetIRIIterateFn(ci),
		g.LoadNavigationIterateFn(ctx, ci, entries),
		g.UpdateIterateIntFn(ci, tx, "state", int(byState), int(setState)),
	)

	err := p.Iterate(ctx).EachValue(nil, ci.Next)
	log.Info().Msg("calling iterate complete")

	if err != nil {
		log.Fatal().Err(err).Msg("blamo")
	}

	if err := g.GraphStore.ApplyTransaction(tx); err != nil {
		log.Error().Err(err).Msg("tx failed")
	}

	it := g.GraphStore.QuadsAllIterator()
	defer it.Close()
	fmt.Println("\nquads:")
	for it.Next(ctx) {
		fmt.Println(g.GraphStore.Quad(it.Result()))
	}
	return nil
}

func (g *CrawlGraph) GetIRIIterateFn(c *iterateCtx) iterateFn {
	return func(value quad.Value) {
		log.Info().Msgf("IN FIRST ASSIGNING IRI")
		nativeValue := quad.NativeOf(value) // this converts RDF values to normal Go types
		c.IRI, _ = nativeValue.(quad.IRI)
		c.Value = value
		c.Next(value)
	}

}

func (g *CrawlGraph) UpdateIterateIntFn(c *iterateCtx, tx *graph.Transaction, field string, from, to int) iterateFn {
	return func(value quad.Value) {
		log.Info().Msgf("IN UPDATE READING IRI")
		log.Info().Msgf("UPDATING %s", c.IRI.String())
		tx.RemoveQuad(quad.Make(c.IRI, quad.IRI(field), quad.Int(from), nil))
		tx.AddQuad(quad.Make(c.IRI, quad.IRI(field), quad.Int(to), nil))
		c.Next(value)
	}
}

func (g *CrawlGraph) LoadNavigationIterateFn(ctx context.Context, c *iterateCtx, entries map[string]*browserker.Navigation) iterateFn {
	return func(value quad.Value) {
		log.Info().Msgf("LOAD OBJ")
		entry := &browserker.Navigation{}
		g.cfg.LoadTo(ctx, g.GraphStore, entry, c.IRI)
		entries[c.IRI.String()] = entry
		c.Next(value)
	}
}

// Close the graph and request stores
func (g *CrawlGraph) Close() error {
	err1 := g.RequestStore.Close()
	err2 := Close(g.GraphStore)

	if err2 != nil {
		// print request store unable to close but just return the graph error
		if err1 != nil {
			log.Error().Err(err1).Msg("failed to close request store")
		}
		return err2
	}

	return err1
}
