package scanner

import (
	"gitlab.com/browserker/browserker"
	"gitlab.com/browserker/store"
)

type Browserk struct {
	cfg         *browserker.Config
	attackGraph store.Storer
	crawlGraph  store.Storer
}

func New(cfg *browserker.Config, attack, crawl store.Storer) *Browserk {
	return &BrowserkScanner{cfg: cfg, attackGraph: attack, crawlGraph: crawl}
}

func (b *Browserk) Init() error {
	if err := b.attackGraph.Init(); err != nil {
		return err
	}
	if err := b.crawlGraph.Init(); err != nil {
		return err
	}

	return nil
}
func (b *Browserk) Start() error {
	return nil
}
