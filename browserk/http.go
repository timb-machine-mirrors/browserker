package browserk

import "github.com/wirepair/gcd/gcdapi"

// revive:disable:var-naming

// HTTPRequest contains all information regarding a network request
type HTTPRequest struct {
	RequestId        string                   `json:"requestId"`                  // Request identifier.
	LoaderId         string                   `json:"loaderId"`                   // Loader identifier. Empty string if the request is fetched from worker.
	DocumentURL      string                   `json:"documentURL"`                // URL of the document this request is loaded for.
	Request          *gcdapi.NetworkRequest   `json:"request"`                    // Request data.
	Timestamp        float64                  `json:"timestamp"`                  // Timestamp.
	WallTime         float64                  `json:"wallTime"`                   // Timestamp.
	Initiator        *gcdapi.NetworkInitiator `json:"initiator"`                  // Request initiator.
	RedirectResponse *gcdapi.NetworkResponse  `json:"redirectResponse,omitempty"` // Redirect response data.
	Type             string                   `json:"type,omitempty"`             // Type of this resource. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	FrameId          string                   `json:"frameId,omitempty"`          // Frame identifier.
	HasUserGesture   bool                     `json:"hasUserGesture,omitempty"`   // Whether the request is initiated by a user gesture. Defaults to false.
}

// HTTPResponse contains all information regarding a network response
type HTTPResponse struct {
	RequestId string                  `json:"requestId"`         // Request identifier.
	LoaderId  string                  `json:"loaderId"`          // Loader identifier. Empty string if the request is fetched from worker.
	Timestamp float64                 `json:"timestamp"`         // Timestamp.
	Type      string                  `json:"type"`              // Resource type. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	Response  *gcdapi.NetworkResponse `json:"response"`          // Response data.
	FrameId   string                  `json:"frameId,omitempty"` // Frame identifier.
	request   HTTPRequest
}

// InterceptedHTTPRequest contains all information regarding an intercepted request
type InterceptedHTTPRequest struct {
	Original  *HTTPRequest
	RequestId string `json:"requestId"` // Each request the page makes will have a unique id.
	//Request             *NetworkRequest            `json:"request"`                       // The details of the request.
	FrameId             string                     `json:"frameId"`                       // The id of the frame that initiated the request.
	ResourceType        string                     `json:"resourceType"`                  // How the requested resource will be used. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	ResponseErrorReason string                     `json:"responseErrorReason,omitempty"` // Response error if intercepted at response stage. enum values: Failed, Aborted, TimedOut, AccessDenied, ConnectionClosed, ConnectionReset, ConnectionRefused, ConnectionAborted, ConnectionFailed, NameNotResolved, InternetDisconnected, AddressUnreachable, BlockedByClient, BlockedByResponse
	ResponseStatusCode  int                        `json:"responseStatusCode,omitempty"`  // Response code if intercepted at response stage.
	ResponseHeaders     []*gcdapi.FetchHeaderEntry `json:"responseHeaders,omitempty"`     // Response headers if intercepted at the response stage.
	NetworkId           string                     `json:"networkId,omitempty"`           // If the intercepted request had a corresponding Network.requestWillBeSent event fired for it, then this networkId will be the same as the requestId present in the requestWillBeSent event.
	Modified            *HTTPModifiedRequest
}

// HTTPModifiedRequest contains the modified http request data
type HTTPModifiedRequest struct {
	RequestId string                     `json:"requestId"`          // An id the client received in requestPaused event.
	Url       string                     `json:"url,omitempty"`      // If set, the request url will be modified in a way that's not observable by page.
	Method    string                     `json:"method,omitempty"`   // If set, the request method is overridden.
	PostData  string                     `json:"postData,omitempty"` // If set, overrides the post data in the request.
	Headers   []*gcdapi.FetchHeaderEntry `json:"headers,omitempty"`  // If set, overrides the request headers.
}

// HTTPModifiedResponse contains the modified http response data
type HTTPModifiedResponse struct {
	RequestId             string                     `json:"requestId"`
	ResponseCode          int                        `json:"responseCode"`                    // An HTTP response code.
	ResponseHeaders       []*gcdapi.FetchHeaderEntry `json:"responseHeaders,omitempty"`       // Response headers.
	BinaryResponseHeaders string                     `json:"binaryResponseHeaders,omitempty"` // Alternative way of specifying response headers as a \0-separated series of name: value pairs. Prefer the above method unless you need to represent some non-UTF8 values that can't be transmitted over the protocol as text.
	Body                  string                     `json:"body,omitempty"`                  // A response body.
	ResponsePhrase        string                     `json:"responsePhrase,omitempty"`        // A textual representation of responseCode. If absent, a standard phrase matching responseCode is used.
}
