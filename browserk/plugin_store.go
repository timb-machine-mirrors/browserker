package browserk

// UniqueFor determines if a plugin event is unique for host/path/query etc
type UniqueFor int8

const (
	UniqueHost UniqueFor = iota
	UniquePath
	UniqueQuery
	UniqueParms
	UniqueMetadata
)

// PluginStorer handles uniqueness and state for plugins
type PluginStorer interface {
	Init() error
	AddEvent(evt *PluginEvent)
	IsUnique(evt *PluginEvent, uniqueFor UniqueFor) bool
	Close() error
}
