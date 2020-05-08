package browserk

// Scope of requests
type Scope int8

const (
	// In scope (we attack)
	In Scope = iota + 1
	// Out of scope (we do not attack, but access)
	Out
	// Excluded from scope (we do not access or attack)
	Excluded
)

// ScopeService checks if a url is in scope
type ScopeService interface {
	Check(url string) Scope
}
