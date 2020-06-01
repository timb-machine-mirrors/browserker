package browserk

import "context"

// PluginServicer does what it says
type PluginServicer interface {
	Name() string
	Init(ctx context.Context) error
	Register(plugin Plugin)
	Unregister(plugin Plugin)
	DispatchEvent(evt *PluginEvent)
	Store() PluginStorer
}
