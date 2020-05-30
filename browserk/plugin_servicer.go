package browserk

import "context"

// PluginServicer does what it says
type PluginServicer interface {
	Init(ctx context.Context) error
	Register(plugin Plugin)
	Unregister(pluginID string)
	DispatchEvent(evt *PluginEvent)
}
