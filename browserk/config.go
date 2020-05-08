package browserk

type Credentials struct {
	Username string
	Password string
	Email    string
}

// AuthType defines how we are going to authenticate
type AuthType int8

const (
	Script AuthType = iota
	Raw
)

type Config struct {
	URL          string
	AllowedURLs  []string // considered 'in scope' for testing/access
	IgnoredURLs  []string // will access, but not report/run tests against (this is the default for non AllowedURLs)
	ExcludedURLs []string // will not allow access to
	DataPath     string
	AuthScript   string
	AuthType     AuthType
	Credentials  *Credentials
	NumBrowsers  int
}
