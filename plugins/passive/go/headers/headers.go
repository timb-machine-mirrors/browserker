package headers

import "gitlab.com/browserker/browserk"

type HeaderPlugin struct {
}

func New() *HeaderPlugin {
	return &HeaderPlugin{}
}

// Name
func (h *HeaderPlugin) Name() string {
	return "HeaderPlugin"
}

// ID
func (h *HeaderPlugin) ID() string {
	return "BR-P-0001"
}

func (h *HeaderPlugin) Config() *browserk.PluginConfig {
	return nil
}

func (h *HeaderPlugin) Options() *browserk.PluginOpts {
	return nil
}

func (h *HeaderPlugin) Register() error {
	return nil
}
func (h *HeaderPlugin) Unregister() error {
	return nil
}
func (h *HeaderPlugin) Ready(browser *browserk.Browser) (bool, error) {
	return false, nil
}

func (h *HeaderPlugin) OnEvent(evt *browserk.PluginEvent, data []byte) {

}
