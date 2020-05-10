package crawler

import (
	"context"
	"time"

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
func (b *BrowserkCrawler) Process(ctx *browserk.Context, browser browserk.Browser, entry *browserk.Navigation, isFinal bool) (*browserk.NavigationResult, []*browserk.Navigation, error) {
	errors := make([]error, 0)
	startURL, err := browser.GetURL()
	if err != nil {
		errors = append(errors, err)
	}
	startCookies, err := browser.GetCookies()
	//clear out storage events
	browser.GetStorageEvents()

	result := &browserk.NavigationResult{
		ID:           nil,
		NavigationID: entry.ID,
		StartURL:     startURL,
		Errors:       errors,
		Cookies:      startCookies,
	}

	navCtx, cancel := context.WithTimeout(ctx.Ctx, time.Second*15)
	defer cancel()
	beforeAction := time.Now()
	_, err = browser.ExecuteAction(navCtx, entry.Action)
	if err != nil {
		result.WasError = true
		return result, nil, err
	}

	b.buildResult(result, beforeAction, browser)

	potentialNavs := make([]*browserk.Navigation, 0)
	if isFinal {
		navCtx, cancel := context.WithTimeout(ctx.Ctx, time.Second*15)
		defer cancel()
		potentialNavs = b.FindNewNav(navCtx, result, browser)
	}
	return result, potentialNavs, nil
}

func (b *BrowserkCrawler) buildResult(result *browserk.NavigationResult, start time.Time, browser browserk.Browser) {
	messages, err := browser.GetMessages()
	result.AddError(err)
	result.Messages = browserk.MessagesAfterRequestTime(messages, start)
	result.MessageCount = len(result.Messages)
	dom, err := browser.GetDOM()
	result.AddError(err)
	result.DOM = dom
	endURL, err := browser.GetURL()
	result.AddError(err)
	result.EndURL = endURL
	cookies, err := browser.GetCookies()
	result.AddError(err)
	result.Cookies = browserk.DiffCookies(result.Cookies, cookies)
	result.StorageEvents = browser.GetStorageEvents()
}

// FindNewNav potentials
func (b *BrowserkCrawler) FindNewNav(ctx context.Context, result *browserk.NavigationResult, browser browserk.Browser) []*browserk.Navigation {

	return nil
}
