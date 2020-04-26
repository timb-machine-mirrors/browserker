package browserker

type ExecutionType int8

const (
	ONCE ExecutionType = iota
	ONCE_PATH
	ONCE_PER_PAGE
	ALWAYS
	ONLY_MIME
	ONLY_INJECTION
)

type PluginOpts struct {
	WriteResponses bool
	WriteRequests  bool
	WriteJS        bool
	ReadResponses  bool
	ReadRequests   bool
	ListenStorage  bool
	ListenCookies  bool
	ListenURL      bool
	ListenJS       bool
	ExecutionType  ExecutionType
	Mimes          []string
	Injections     []string
}

type PluginCheck struct {
	CWE         string
	Name        string
	Description string
	CheckID     string
}

type PluginConfig struct {
	Class    string
	Plugin   string
	Language string
	ID       int
}

type Plugin interface {
	Name() string
	ID() string
	Config() *PluginConfig
	Register() error
	Unregister() error

	OnRequest(requestID string, data []byte)
}
