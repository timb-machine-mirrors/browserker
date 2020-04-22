package scanner

import (
	"gitlab.com/browserker/browserker"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/report"
	"gitlab.com/browserker/store"
)

type Browserk struct {
	cfg         *browserker.Config
	attackGraph store.Storer
	crawlGraph  store.Storer
	reporter    browserker.Reporter
	browsers    browser.Browser
}

func New(cfg *browserker.Config, attack, crawl store.Storer) *Browserk {
	return &Browserk{cfg: cfg, attackGraph: attack, crawlGraph: crawl, reporter: report.New()}
}

func (b *Browserk) SetReporter(reporter browserker.Reporter) *Browserk {
	b.reporter = reporter
	return b
}

func (b *Browserk) Init() error {
	if err := b.attackGraph.Init(); err != nil {
		return err
	}

	if err := b.crawlGraph.Init(); err != nil {
		return err
	}

	// TODO: load browsers
	leaser := browser.NewLocalLeaser()
	b.browsers = browser.NewGCDBrowserPool(b.cfg.NumBrowsers, leaser)

	return nil
}

func (b *Browserk) Start() error {
	return nil
}

func (b *Browserk) Stop() error {
	return nil
}
