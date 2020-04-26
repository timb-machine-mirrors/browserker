package browserker

import "context"

type CrawlStep func(ctx context.Context, browser Browser) (next CrawlStep, err error)

type Crawler interface {
	Init() error
	Start() error
	GoTo(next CrawlStep) (CrawlStep, error)
	Stop() error
}
