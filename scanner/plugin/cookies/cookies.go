package cookies

import "gitlab.com/browserker/browserk"

type Plugin struct {
	service browserk.PluginServicer
}

func New(service browserk.PluginServicer) *Plugin {
	return &Plugin{service: service}
}

// Name of the plugin
func (h *Plugin) Name() string {
	return "CookiePlugin"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-P-0001"
}

// Config for this plugin
func (h *Plugin) Config() *browserk.PluginConfig {
	return nil
}

// Options for the plugin manager to take into consideration when dispatching
func (h *Plugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenCookies: true,
	}
}

// Ready to attack
func (h *Plugin) Ready(browser *browserk.Browser) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {

}
