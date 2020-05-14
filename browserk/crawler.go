package browserk

import "context"

// Crawler service
type Crawler interface {
	Init() error
	Start(ctx context.Context) error
	Stop() error
}
