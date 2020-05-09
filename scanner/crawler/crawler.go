package crawler

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"gitlab.com/browserker/browserk"
)

type BrowserkCrawler struct {
	cfg *browserk.Config
}

func New(cfg *browserk.Config) *BrowserkCrawler {
	return &BrowserkCrawler{cfg: cfg}
}

func (b *BrowserkCrawler) Init() error {
	return nil
}

// Process the next navigation entry
func (b *BrowserkCrawler) Process(ctx *browserk.Context, browser browserk.Browser, entry *browserk.Navigation, isFinal bool) ([]*browserk.NavigationResult, []*browserk.Navigation, error) {
	switch entry.Action.Type {
	case browserk.ActLoadURL:
		if err := browser.Navigate(ctx.Ctx, string(entry.Action.Input)); err != nil {
			return nil, nil, err
		}
		spew.Dump(browser.GetMessages())
	}
	return nil, nil, nil
}

// FindNewNav potentials
func (b *BrowserkCrawler) FindNewNav(ctx context.Context, browser browserk.Browser) []*browserk.Navigation {
	return nil
}
