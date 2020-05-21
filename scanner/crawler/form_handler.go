package crawler

import "gitlab.com/browserker/browserk"

type CrawlerFormHandler struct {
}

func NewCrawlerFormHandler() *CrawlerFormHandler {
	return &CrawlerFormHandler{}
}

// Init the form filler
func Init() error {
	return nil
}

// Fill the form with context relevant data
func (c *CrawlerFormHandler) Fill(form *browserk.HTMLFormElement) {
	labels := make(map[string]*browserk.HTMLElement)
	// iterate once to grab labels/context
	for _, ele := range form.ChildElements {
		if forInput, ok := ele.Attributes["for"]; ok && ele.Type == browserk.LABEL {
			labels[forInput] = ele
		}
	}
}
