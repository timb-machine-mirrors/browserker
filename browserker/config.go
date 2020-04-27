package browserker

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
	AllowedHosts []string
	IgnoredHosts []string
	DataPath     string
	AuthScript   string
	AuthType     AuthType
	Credentials  *Credentials
	NumBrowsers  int
}
