package browserker

type ExecutionType int8

// revive:disable:var-naming

const (
	ONCE ExecutionType = iota
	ONCE_PATH
	ONCE_PER_PAGE
	ALWAYS
	ONLY_MIME
	ONLY_INJECTION
)

type PluginEvent int8

const (
	DOCUMENT_REQUEST PluginEvent = iota
	HTTP_REQUEST
	HTTP_RESPONSE
	WEBSOCKET_REQUEST
	WEBSOCKET_RESPONSE
	URL
	JS_RESPONSE
	STORAGE
	COOKIE
)

type PluginOpts struct {
	IsolatedRequests bool          // Initiates it's own requests, isolated from a crawl state
	WriteResponses   bool          // writes/injects into http/websocket responses
	WriteRequests    bool          // writes/injects into http/websocket responses
	WriteJS          bool          // writes/injects JS into the browser
	ReadResponses    bool          // reads http/websocket responses
	ReadRequests     bool          // reads http/websocket requests
	ListenStorage    bool          // listens for local/sessionStorage write/read events
	ListenCookies    bool          // listens for cookie write events
	ListenURL        bool          // listens for URL change/updates
	ListenJS         bool          // listens to JS events
	ExecutionType    ExecutionType // How often/when this plugin executes
	Mimes            []string      // list of mime types this plugin will execute on if ExecutionType = ONLY_INJECTION
	Injections       []string      // list of injection points this plugin will execute on
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
	Options() *PluginOpts
	Register() error
	Unregister() error
	Ready(browser Browser) (bool, error) // ready for injection or whatever, ret true if injected
	OnEvent(evt PluginEvent, data []byte)
}
