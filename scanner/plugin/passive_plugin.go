package plugin

type PassiveCheck struct {
	CWE         string
	Name        string
	Description string
	CheckID     string
}
type PassiveConfig struct {
	Class    string
	Plugin   string
	Language string
	ID       int
	Checks   []*PassiveCheck
}

type PassivePlugin interface {
	Name() string
	ID() string
	Config() *PassiveConfig
	Register() error
	Unregister() error

	OnWebSocketRequest(data []byte)
	OnRequest(requestID string, data []byte)
}
