package browserker

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
	AddNavigation(nav *Navigation) error
}
