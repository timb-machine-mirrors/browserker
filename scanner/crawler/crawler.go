package crawler

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

// BrowserkCrawler crawls a site
type BrowserkCrawler struct {
	cfg *browserk.Config
}

// New crawler for a site
func New(cfg *browserk.Config) *BrowserkCrawler {
	return &BrowserkCrawler{cfg: cfg}
}

// Init the crawler, if necessary
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

	//clear out storage and console events before executing our action
	browser.GetStorageEvents()
	browser.GetConsoleEvents()

	result := &browserk.NavigationResult{
		ID:           nil,
		NavigationID: entry.ID,
		StartURL:     startURL,
		Errors:       errors,
		Cookies:      startCookies,
	}

	// execute the action
	navCtx, cancel := context.WithTimeout(ctx.Ctx, time.Second*15)
	defer cancel()
	beforeAction := time.Now()
	_, result.CausedLoad, err = browser.ExecuteAction(navCtx, entry.Action)
	if err != nil {
		result.WasError = true
		return result, nil, err
	}

	// capture results
	b.buildResult(result, beforeAction, browser)

	// find new potential navigation entries (if isFinal)
	potentialNavs := make([]*browserk.Navigation, 0)
	if isFinal {
		potentialNavs = b.FindNewNav(entry, browser)
	}
	return result, potentialNavs, nil
}

// buildResult captures various data points after we executed an Action
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
	result.ConsoleEvents = browser.GetConsoleEvents()
	result.Hash()
}

// FindNewNav potentials TODO: get navigation entry metadata (is vuejs/react etc) to be more specific
func (b *BrowserkCrawler) FindNewNav(entry *browserk.Navigation, browser browserk.Browser) []*browserk.Navigation {
	navs := make([]*browserk.Navigation, 0)
	// Pull out forms (highest priority)
	formElements, err := browser.FindForms()
	if err != nil {
		log.Info().Err(err).Msg("error while extracting forms")
	} else if formElements != nil && len(formElements) > 0 {
		log.Debug().Int("form_count", len(formElements)).Msg("found new forms")
		for _, form := range formElements {
			navs = append(navs, browserk.NewNavigationFromForm(entry, browserk.TrigCrawler, form))
		}
	}

	bElements, err := browser.FindElements("button")
	if err != nil {
		log.Info().Err(err).Msg("error while extracting buttons")
	} else if bElements != nil && len(bElements) > 0 {
		log.Debug().Int("link_count", len(bElements)).Msg("found buttons")
		for _, b := range bElements {
			navs = append(navs, browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, b, browserk.ActLeftClick))
		}
	}

	// pull out links (lower priority)
	aElements, err := browser.FindElements("a")
	if err != nil {
		log.Info().Err(err).Msg("error while extracting links")
	} else if aElements != nil && len(aElements) > 0 {
		log.Debug().Int("link_count", len(aElements)).Msg("found links")
		for _, a := range aElements {
			navs = append(navs, browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, a, browserk.ActLeftClick))
		}
	}

	// todo pull out additional clickable/whateverable elements
	spew.Dump(navs)
	return navs
}
