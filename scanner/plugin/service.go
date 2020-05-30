package plugin

import (
	"context"
	"sync"

	"gitlab.com/browserker/browserk"
)

// Service of plugins
type Service struct {
	cfg         *browserk.Config
	ctx         context.Context
	pluginStore browserk.PluginStorer
	eventCh     chan *browserk.PluginEvent

	pluginLock *sync.RWMutex
	plugins    map[string]browserk.Plugin
}

// New plugin manager
func New(cfg *browserk.Config, pluginStore browserk.PluginStorer) *Service {
	return &Service{
		cfg:         cfg,
		pluginStore: pluginStore,
		eventCh:     make(chan *browserk.PluginEvent),
		plugins:     make(map[string]browserk.Plugin),
		pluginLock:  &sync.RWMutex{},
	}
}

func (s *Service) Register(plugin browserk.Plugin) {
	s.pluginLock.Lock()
	s.plugins[plugin.ID()] = plugin
	s.pluginLock.Unlock()
}

func (s *Service) Unregister(pluginID string) {
	s.pluginLock.Lock()
	delete(s.plugins, pluginID)
	s.pluginLock.Unlock()
}

// Init the plugin manager
func (s *Service) Init(ctx context.Context) error {
	s.ctx = ctx
	go s.listenForEvents()
	// TODO: load plugins

	return nil
}

// DispatchEvent to interested listeners
func (s *Service) DispatchEvent(evt *browserk.PluginEvent) {
	select {
	case <-s.ctx.Done():
		return
	case s.eventCh <- evt:
	}
}

func (s *Service) listenForEvents() {
	for {
		select {
		case evt := <-s.eventCh:
			s.pluginStore.IsUnique(evt)
			switch evt.Type {
			case browserk.EvtHTTPRequest:
			case browserk.EvtHTTPResponse:
			case browserk.EvtInterceptedHTTPRequest:
			case browserk.EvtInterceptedHTTPResponse:
			case browserk.EvtWebSocketRequest:
			case browserk.EvtWebSocketResponse:
			case browserk.EvtURL:
			case browserk.EvtJSResponse:
			case browserk.EvtStorage:
			case browserk.EvtCookie:
			}
		case <-s.ctx.Done():
			return
		}
	}
}
