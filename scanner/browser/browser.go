package browser

import (
	"context"
)

type Browser interface {
	// Load a web page, return the dom string, responses
	Load(ctx context.Context, address, scheme, port string) (err error)
}
