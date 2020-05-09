package mock

import (
	"context"

	"gitlab.com/browserker/browserk"
)

func Context(ctx context.Context) *browserk.Context {
	return &browserk.Context{
		Ctx:      ctx,
		Auth:     nil,
		Scope:    nil,
		Reporter: nil,
		Injector: nil,
		Crawl:    nil,
		Attack:   nil,
	}
}
