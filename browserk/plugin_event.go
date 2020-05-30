package browserk

import "go/parser"

// revive:disable:var-naming

type PluginEventType int8

const (
	EvtDocumentRequest PluginEventType = iota
	EvtHTTPRequest
	EvtHTTPResponse
	EvtInterceptedHTTPRequest
	EvtInterceptedHTTPResponse
	EvtWebSocketRequest
	EvtWebSocketResponse
	EvtURL
	EvtJSResponse
	EvtStorage
	EvtCookie
)

type PluginEvent struct {
	Type      PluginEventType
	URL       string
	Nav       *Navigation
	BCtx      Context
	EventData *PluginEventData
}

type PluginEventData struct {
	HTTPRequest             *HTTPRequest
	HTTPResponse            *HTTPResponse
	InterceptedHTTPRequest  *InterceptedHTTPRequest
	InterceptedHTTPResponse *InterceptedHTTPResponse
	Storage                 *StorageEvent
	Cookie                  *Cookie
}

func HTTPRequestPluginEvent(bctx Context, URL string, nav *Navigation, request *HTTPRequest) *PluginEvent {
	ast.ParseFile(parser.DeclarationErrors)
	evt := newPluginEvent(bctx, URL, nav, EvtHTTPRequest)
	evt.EventData = &PluginEventData{HTTPRequest: request}
	return evt
}

func HTTPResponsePluginEvent(bctx Context, URL string, nav *Navigation, response *HTTPResponse) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtHTTPResponse)
	evt.EventData = &PluginEventData{HTTPResponse: response}
	return evt
}

func InterceptedHTTPRequestPluginEvent(bctx Context, URL string, nav *Navigation, request *InterceptedHTTPRequest) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtInterceptedHTTPRequest)
	evt.EventData = &PluginEventData{InterceptedHTTPRequest: request}
	return evt
}

func InterceptedHTTPResponsePluginEvent(bctx Context, URL string, nav *Navigation, response *InterceptedHTTPResponse) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtInterceptedHTTPResponse)
	evt.EventData = &PluginEventData{InterceptedHTTPResponse: response}
	return evt
}

func StoragePluginEvent(bctx Context, URL string, nav *Navigation, storage *StorageEvent) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtStorage)
	evt.EventData = &PluginEventData{Storage: storage}
	return evt
}

func CookiePluginEvent(bctx Context, URL string, nav *Navigation, cookie *Cookie) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtCookie)
	evt.EventData = &PluginEventData{Cookie: cookie}
	return evt
}

func newPluginEvent(bctx Context, URL string, nav *Navigation, eventType PluginEventType) *PluginEvent {
	return &PluginEvent{
		Type: eventType,
		URL:  URL,
		Nav:  nav,
		BCtx: bctx,
	}
}
