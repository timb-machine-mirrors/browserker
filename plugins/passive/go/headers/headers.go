package headers

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
	"BR-P-0001"
}

func (h *HeaderPlugin) Config() *PluginConfig {
	return nil
}

func (h *HeaderPlugin) Options() *PluginOpts {
	return nil
}

func (h *HeaderPlugin) Register() error {
	return nil
}
func (h *HeaderPlugin) Unregister() error {
	return nil
}
func (h *HeaderPlugin) Ready(browser Browser) (bool, error) {
	return false, nil
}

func (h *HeaderPlugin) OnEvent(evt PluginEvent, data []byte) {

}
