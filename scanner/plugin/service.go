package plugin

import (
	"context"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/plugin/cookies"
	"gitlab.com/browserker/scanner/plugin/headers"
	"gitlab.com/browserker/scanner/plugin/storage"
)

// Service of plugins
type Service struct {
	cfg         *browserk.Config
	ctx         context.Context
	pluginStore browserk.PluginStorer
	eventCh     chan *browserk.PluginEvent

	hostPlugins     *Container
	pathPlugins     *Container
	filePlugins     *Container
	pagePlugins     *Container
	requestPlugins  *Container
	responsePlugins *Container
	alwaysPlugins   *Container
}

// New plugin manager
func New(cfg *browserk.Config, pluginStore browserk.PluginStorer) *Service {
	return &Service{
		cfg:             cfg,
		pluginStore:     pluginStore,
		eventCh:         make(chan *browserk.PluginEvent),
		hostPlugins:     NewContainer(),
		pathPlugins:     NewContainer(),
		filePlugins:     NewContainer(),
		pagePlugins:     NewContainer(),
		requestPlugins:  NewContainer(),
		responsePlugins: NewContainer(),
		alwaysPlugins:   NewContainer(),
	}
}

func (s *Service) Register(plugin browserk.Plugin) {
	plugins := s.getPluginsOfType(plugin.Options().ExecutionType)
	plugins.Add(plugin)
}

func (s *Service) Store() browserk.PluginStorer {
	return s.pluginStore
}

func (s *Service) getPluginsOfType(pluginType browserk.PluginExecutionType) *Container {
	switch pluginType {
	case browserk.ExecOnce:
		return s.hostPlugins
	case browserk.ExecOncePath:
		return s.pathPlugins
	case browserk.ExecOnceFile:
		return s.filePlugins
	case browserk.ExecOncePerPage:
		return s.pagePlugins
	case browserk.ExecPerRequest:
		return s.requestPlugins
	case browserk.ExecAlways:
		return s.alwaysPlugins
	}
	return nil
}

// Unregister the plugin based on type
func (s *Service) Unregister(plugin browserk.Plugin) {
	plugins := s.getPluginsOfType(plugin.Options().ExecutionType)
	plugins.Remove(plugin)
}

// Init the plugin manager
func (s *Service) Init(ctx context.Context) error {
	s.ctx = ctx
	importPlugins(s)
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
			u := s.pluginStore.IsUnique(evt)
			if u.Host() {
				s.hostPlugins.Call(evt)
			}
			if u.Path() {
				s.pathPlugins.Call(evt)
			}
			if u.File() {
				s.filePlugins.Call(evt)
			}
			if u.Page() {
				s.pagePlugins.Call(evt)
			}
			if u.Request() {
				s.requestPlugins.Call(evt)
			}
			if u.Response() {
				s.responsePlugins.Call(evt)
			}
			s.alwaysPlugins.Call(evt)
		case <-s.ctx.Done():
			return
		}
	}
}

func importPlugins(s *Service) {
	s.Register(cookies.New(s))
	s.Register(headers.New(s))
	s.Register(storage.New(s))
}
