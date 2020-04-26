package browserker

type Scanner interface {
	Init() error
	Start() error
	Pause() error
	Stop() error
	Plugins() map[string]Plugin
	Reporter() Reporter
}
