package crawler

import (
	"context"

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

func (b *BrowserkCrawler) Process(ctx context.Context, browser browserk.Browser, entry *browserk.Navigation) ([]*browserk.NavigationResult, []*browserk.Navigation, error) {
	switch entry.Action.Type {
	case browserk.ActLoadURL:
		if err := browser.Load(ctx, string(entry.Action.Input)); err != nil {
			return nil, nil, err
		}
		browser.
	}
	return nil, nil, nil
}

func (b *BrowserkCrawler) FindNewNav(ctx context.Context, browser browserk.Browser) []*browserk.Navigation {
	browser.GetEventListeners(ctx)
}