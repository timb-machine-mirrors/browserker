package browserk

import "context"

// AttackGrapher is a graph based storage system
type AttackGrapher interface {
	Init() error
	Close() error
	AddAttack()
}

// CrawlGrapher is a graph based storage system
type CrawlGrapher interface {
	Init() error
	Close() error
	Find(ctx context.Context, byState, setState NavState, limit int64) [][]*Navigation
	AddNavigation(nav *Navigation) error
	AddNavigations(navs []*Navigation) error
	FailNavigation(navID []byte) error
	AddResult(result *NavigationResult) error
	NavExists(nav *Navigation) bool
	GetNavigation(id []byte) (*Navigation, error)
}
