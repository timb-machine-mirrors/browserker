package browserk

// Scope of requests
type Scope int8

const (
	// In scope
	In Scope = iota + 1
	// Out of scope
	Out
	// Excluded from scope
	Excluded
)

// ScopeService checks if a url is in scope
type ScopeService interface {
	Check(url string) Scope
}
