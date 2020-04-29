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
func (g *CrawlGraph) Find(ctx context.Context, byState, setState browserker.NavState) map[string]*browserker.Navigation {
	//entries := make(map[string]*browserker.Navigation, 0)
	p := cayley.StartPath(g.GraphStore).Has(quad.IRI("state"), quad.Int(byState)) // (quad.IRI("state")) .Out(quad.IRI("origin"))
	//spew.Dump(p)
	log.Info().Msg("calling iterate")
	vs := make([]quad.IRI, 0)

	err := p.Iterate(ctx).EachValue(nil, func(value quad.Value) {
		nativeValue := quad.NativeOf(value) // this converts RDF values to normal Go types
		log.Info().Msgf("in iterate %v", nativeValue)
		iri, _ := nativeValue.(quad.IRI)

		vs = append(vs, iri)
	})
	log.Info().Msg("calling iterate complete")

	if err != nil {
		log.Fatal().Err(err).Msg("blamo")
	}
	tx := cayley.NewTransaction()
	for _, iri := range vs {
		tx.RemoveQuad(quad.Make(iri, quad.IRI("state"), quad.Int(byState), nil))
		tx.AddQuad(quad.Make(iri, quad.IRI("state"), quad.Int(setState), nil))
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
