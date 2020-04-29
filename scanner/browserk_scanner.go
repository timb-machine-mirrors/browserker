package scanner

import (
	"github.com/rs/zerolog/log"

	"gitlab.com/browserker/browserker"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/report"
)

// Browserk is our engine
type Browserk struct {
	cfg         *browserker.Config
	attackGraph browserker.AttackGrapher
	crawlGraph  browserker.CrawlGrapher
	reporter    browserker.Reporter
	browsers    browser.BrowserPool
}

// New engine
func New(cfg *browserker.Config, crawl browserker.CrawlGrapher, attack browserker.AttackGrapher) *Browserk {
	return &Browserk{cfg: cfg, attackGraph: attack, crawlGraph: crawl, reporter: report.New()}
}

// SetReporter overrides the default reporter
func (b *Browserk) SetReporter(reporter browserker.Reporter) *Browserk {
	b.reporter = reporter
	return b
}

// Init the browsers and stores
func (b *Browserk) Init() error {
	log.Logger.Info().Msg("initializing attack graph")
	if err := b.attackGraph.Init(); err != nil {
		return err
	}

	log.Logger.Info().Msg("initializing crawl graph")
	if err := b.crawlGraph.Init(); err != nil {
		return err
	}
	log.Logger.Info().Msg("starting leaser")
	leaser := browser.NewLocalLeaser()
	log.Logger.Info().Msg("leaser started")
	pool := browser.NewGCDBrowserPool(b.cfg.NumBrowsers, leaser)
	b.browsers = pool
	log.Logger.Info().Msg("starting browser pool")
	return pool.Init()
}

// Start the browsers
func (b *Browserk) Start() error {
	return nil //b.browsers.Load(context.Background(), b.cfg.URL)
}

// Stop the browsers
func (b *Browserk) Stop() error {
	return nil
}
