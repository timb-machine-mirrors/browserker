package browser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/browserk/inject"
	"gitlab.com/browserker/browserk/navi"

	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

// Tab is a chromium browser tab we use for instrumentation
type Tab struct {
	g                     *gcd.Gcd
	t                     *gcd.ChromeTarget
	container             *ResponseContainer
	id                    int64
	eleMutex              *sync.RWMutex          // locks our elements when added/removed.
	elements              map[int]*Element       // our map of elements for this tab
	topNodeID             atomic.Value           // the nodeID of the current top level #document
	topFrameID            atomic.Value           // the frameID of the current top level #document
	isNavigatingFlag      atomic.Value           // are we currently navigating (between Page.Navigate -> page.loadEventFired)
	isTransitioningFlag   atomic.Value           // has navigation occurred on the top frame (not due to Navigate() being called)
	debug                 bool                   // for debug printing
	nodeChange            chan *NodeChangeEvent  // for receiving node change events from tab_subscribers
	navigationCh          chan int               // for receiving navigation complete messages while isNavigating is true
	docUpdateCh           chan struct{}          // for receiving document update completion while isNavigating is true
	crashedCh             chan string            // the chrome tab crashed with a reason
	exitCh                chan struct{}          // for when we close the tab, kill go routines
	shutdown              atomic.Value           // have we already shut down
	disconnectedHandler   TabDisconnectedHandler // called with reason the chrome tab was disconnected from the debugger service
	navigationTimeout     time.Duration          // amount of time to wait before failing navigation
	elementTimeout        time.Duration          // amount of time to wait for element readiness
	stabilityTimeout      time.Duration          // amount of time to give up waiting for stability
	stableAfter           time.Duration          // amount of time of no activity to consIDer the DOM stable
	lastNodeChangeTimeVal atomic.Value           // timestamp of when the last node change occurred atomic because multiple go routines will modify
	domChangeHandler      DomChangeHandlerFunc   // allows the caller to be notified of DOM change events.
}

// NewTab to use
func NewTab(ctx context.Context, gcdBrowser *gcd.Gcd, tab *gcd.ChromeTarget) *Tab {
	t := &Tab{
		t:            tab,
		container:    NewResponseContainer(),
		crashedCh:    make(chan string),
		exitCh:       make(chan struct{}),
		navigationCh: make(chan int),
	}
	t.id = browserk.GetBrowserID()
	t.g = gcdBrowser
	t.eleMutex = &sync.RWMutex{}
	t.elements = make(map[int]*Element)
	t.nodeChange = make(chan *NodeChangeEvent)
	t.navigationCh = make(chan int, 1)  // for signaling navigation complete
	t.docUpdateCh = make(chan struct{}) // wait for documentUpdate to be called during navigation
	t.crashedCh = make(chan string)     // reason the tab crashed/was disconnected.
	t.exitCh = make(chan struct{})
	t.navigationTimeout = 30 * time.Second // default 30 seconds for timeout
	t.elementTimeout = 5 * time.Second     // default 5 seconds for waiting for element.
	t.stabilityTimeout = 2 * time.Second   // default 2 seconds before we give up waiting for stability
	t.stableAfter = 300 * time.Millisecond // default 300 ms for consIDering the DOM stable
	t.domChangeHandler = nil

	t.disconnectedHandler = t.defaultDisconnectedHandler
	t.subscribeBrowserEvents(ctx)
	return t
}

// SetDisconnectedHandler so caller can trap when the debugger was disconnected/crashed.
func (t *Tab) SetDisconnectedHandler(handlerFn TabDisconnectedHandler) {
	t.disconnectedHandler = handlerFn
}

func (t *Tab) defaultDisconnectedHandler(tab *Tab, reason string) {
	log.Debug().Msgf("tab %s tabID: %s", reason, tab.t.Target.Id)
}

// Close the exit channel
func (t *Tab) Close() {
	close(t.exitCh)
}

// Navigate capture network traffic and take screen shot of DOM and image
func (t *Tab) Navigate(ctx context.Context, url string) error {
	navParams := &gcdapi.PageNavigateParams{Url: url, TransitionType: "typed"}
	_, _, errText, err := t.t.Page.NavigateWithParams(navParams)
	if err != nil {
		return err
	}

	if errText != "" {
		return errors.Wrap(ErrNavigating, errText)
	}

	//log.Ctx(ctx).Info().Str("url", url).Str("err_text", errText).Msg("navigating complete")
	err = t.WaitReady(ctx, time.Second*9)
	//log.Ctx(ctx).Info().Msg("wait ready returned")
	return err
}

// ID of this browser (tab)
func (t *Tab) ID() int64 {
	return t.id
}

func (t *Tab) Find(ctx context.Context, finder navi.Find) (*navi.Element, error) {
	return nil, nil
}

func (t *Tab) Instrument(opt *browserk.BrowserOpts) error {
	return nil
}

func (t *Tab) InjectBefore(ctx context.Context, inject inject.Injector) error {
	return nil
}

func (t *Tab) InjectAfter(ctx context.Context, inject inject.Injector) ([]byte, error) {
	return nil, nil
}

func (t *Tab) GetResponses() (map[int64]*browserk.HTTPResponse, error) {
	return nil, nil
}

func (t *Tab) GetRequest() (*browserk.HTTPRequest, error) {
	return nil, nil
}

func (t *Tab) Execute(ctx context.Context, act map[int]*browserk.Action) error {
	return nil
}

func (t *Tab) ExecuteSingle(ctx context.Context, act *browserk.Action) error {
	return nil
}

// InjectJS only caller knows what the response type will be so return an interface{}
// caller must type check to whatever they expect
func (t *Tab) InjectJS(inject string) (interface{}, error) {
	params := &gcdapi.RuntimeEvaluateParams{
		Expression:            inject,
		ObjectGroup:           "browserker",
		IncludeCommandLineAPI: false,
		Silent:                true,
		ReturnByValue:         true,
		GeneratePreview:       false,
		UserGesture:           false,
		AwaitPromise:          false,
		ThrowOnSideEffect:     false,
		Timeout:               1000,
	}
	r, exp, err := t.t.Runtime.EvaluateWithParams(params)
	if err != nil {
		return nil, err
	}
	if exp != nil {
		log.Warn().Err(err).Msg("failed to inject script")
	}

	return r.Value, nil
}

// GetURL by looking at the navigation history
func (t *Tab) GetURL(ctx context.Context) string {
	_, entries, err := t.t.Page.GetNavigationHistory()
	if err != nil || len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Url
}

// WaitReady waits for the page to load, DOM to be stable, and no network traffic in progress
func (t *Tab) WaitReady(ctx context.Context, stableAfter time.Duration) error {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	navTimer := time.After(45 * time.Second)
	//log.Ctx(ctx).Info().Msg("waiting for nav to complete")
	// wait navigation to complete.
	select {
	case <-navTimer:
		return ErrNavigationTimedOut
	case <-ctx.Done():
		return ctx.Err()
	case <-t.exitCh:
		return errors.New("exiting")
	case reason := <-t.crashedCh:
		return errors.Wrap(ErrTabCrashed, reason)
	case <-t.navigationCh:
	}

	stableTimer := time.After(5 * time.Second)

	// wait for DOM & network stability
	//log.Ctx(ctx).Info().Msg("waiting for DOM & network stability")
	for {
		select {
		case reason := <-t.crashedCh:
			return errors.Wrap(ErrTabCrashed, reason)
		case <-ctx.Done():
			return ctx.Err()
		case <-t.exitCh:
			return ErrTabClosing
		case <-stableTimer:
			log.Ctx(ctx).Info().Msg("stability timed out")
			return ErrTimedOut
		case <-ticker.C:
			if changeTime, ok := t.lastNodeChangeTimeVal.Load().(time.Time); ok {
				//log.Info().Int32("requests", t.container.GetRequests()).Msgf("tick %s", time.Now().Sub(changeTime))
				if time.Now().Sub(changeTime) >= stableAfter && t.container.GetRequests() == 0 {
					// times up, should be stable now
					log.Ctx(ctx).Info().Msg("stable")
					return nil
				}
			}
		}
	}
}

// SetNavigationTimeout to wait in seconds for navigations before giving up, default is 30 seconds
func (t *Tab) SetNavigationTimeout(timeout time.Duration) {
	t.navigationTimeout = timeout
}

// SetElementWaitTimeout to wait in seconds for ele.WaitForReady() before giving up, default is 5 seconds
func (t *Tab) SetElementWaitTimeout(timeout time.Duration) {
	t.elementTimeout = timeout
}

// SetStabilityTimeout to wait for WaitStable() to return, default is 2 seconds.
func (t *Tab) SetStabilityTimeout(timeout time.Duration) {
	t.stabilityTimeout = timeout
}

// SetStabilityTime to wait for no node changes before we consIDer the DOM stable.
// Note that stability timeout will fire if the DOM is constantly changing.
// The deafult stableAfter is 300 ms.
func (t *Tab) SetStabilityTime(stableAfter time.Duration) {
	t.stableAfter = stableAfter
}

func (t *Tab) setIsNavigating(set bool) {
	t.isNavigatingFlag.Store(set)
}

// IsNavigating answers if we currently navigating
func (t *Tab) IsNavigating() bool {
	if flag, ok := t.isNavigatingFlag.Load().(bool); ok {
		return flag
	}
	return false
}

func (t *Tab) setIsTransitioning(set bool) {
	t.isTransitioningFlag.Store(set)
}

// IsTransitioning returns true if we are transitioning to a new page. This is not set when Navigate is called.
func (t *Tab) IsTransitioning() bool {
	if flag, ok := t.isTransitioningFlag.Load().(bool); ok {
		return flag
	}
	return false
}

func (t *Tab) setTopFrameID(topFrameID string) {
	t.topFrameID.Store(topFrameID)
}

// GetTopFrameID return the top frame ID of this tab
func (t *Tab) GetTopFrameID() string {
	if frameID, ok := t.topFrameID.Load().(string); ok {
		return frameID
	}
	return ""
}

func (t *Tab) setTopNodeID(nodeID int) {
	t.topNodeID.Store(nodeID)
}

// GetTopNodeID returns the current top node ID of this Tab.
func (t *Tab) GetTopNodeID() int {
	if topNodeID, ok := t.topNodeID.Load().(int); ok {
		return topNodeID
	}
	return -1
}

// DidNavigationFail uses an undocumented method of determining if chromium failed to load
// a page due to DNS or connection timeouts.
func (t *Tab) DidNavigationFail() (bool, string) {
	// if loadTimeData doesn't exist, or we get a js error, this means no error occurred.
	rro, err := t.EvaluateScript("loadTimeData.data_.errorCode")
	if err != nil {
		return false, ""
	}

	if val, ok := rro.Value.(string); ok {
		return true, val
	}

	return false, ""
}

// EvaluateScript in the global context.
func (t *Tab) EvaluateScript(scriptSource string) (*gcdapi.RuntimeRemoteObject, error) {
	return t.evaluateScript(scriptSource, false)
}

// EvaluatePromiseScript in the global context.
func (t *Tab) EvaluatePromiseScript(scriptSource string) (*gcdapi.RuntimeRemoteObject, error) {
	return t.evaluateScript(scriptSource, true)
}

// evaluateScript in the global context.
func (t *Tab) evaluateScript(scriptSource string, awaitPromise bool) (*gcdapi.RuntimeRemoteObject, error) {
	params := &gcdapi.RuntimeEvaluateParams{
		Expression:            scriptSource,
		ObjectGroup:           "browserker",
		IncludeCommandLineAPI: false,
		Silent:                true,
		ReturnByValue:         true,
		GeneratePreview:       false,
		UserGesture:           false,
		AwaitPromise:          awaitPromise,
		ThrowOnSideEffect:     false,
		Timeout:               1000,
	}
	r, exp, err := t.t.Runtime.EvaluateWithParams(params)
	if err != nil {
		return nil, err
	}
	if exp != nil {
		log.Warn().Err(err).Msg("failed to inject script")
	}

	return r, nil
}

// NavigationHistory the current navigation index, history entries or error
func (t *Tab) NavigationHistory() (int, []*gcdapi.PageNavigationEntry, error) {
	return t.t.Page.GetNavigationHistory()
}

// Reload the page injecting evalScript to run on load. set ignoreCache to true
// to have it act like ctrl+f5.
func (t *Tab) Reload(ignoreCache bool, evalScript string) error {
	_, err := t.t.Page.Reload(ignoreCache, evalScript)
	return err
}

// Forward the next navigation entry from the history and navigates to it.
// Returns error if we could not find the next entry or navigation failed
func (t *Tab) Forward() error {
	next, err := t.ForwardEntry()
	if err != nil {
		return err
	}
	_, err = t.t.Page.NavigateToHistoryEntry(next.Id)
	return err
}

// ForwardEntry the next entry in our navigation history for this tab.
func (t *Tab) ForwardEntry() (*gcdapi.PageNavigationEntry, error) {
	idx, entries, err := t.NavigationHistory()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(entries); i++ {
		if idx < entries[i].Id {
			return entries[i], nil
		}
	}
	return nil, &InvalidNavigationErr{Message: "Unable to navigate forward as we are on the latest navigation entry"}
}

// Back the previous navigation entry from the history and navigates to it.
// Returns error if we could not find the previous entry or navigation failed
func (t *Tab) Back() error {
	prev, err := t.BackEntry()
	if err != nil {
		return err
	}
	_, err = t.t.Page.NavigateToHistoryEntry(prev.Id)
	return err
}

// BackEntry the previous entry in our navigation history for this tab.
func (t *Tab) BackEntry() (*gcdapi.PageNavigationEntry, error) {
	idx, entries, err := t.NavigationHistory()
	if err != nil {
		return nil, err
	}

	for i := len(entries); i > 0; i-- {
		if idx < entries[i].Id {
			return entries[i], nil
		}
	}
	return nil, &InvalidNavigationErr{Message: "Unable to navigate backward as we are on the first navigation entry"}
}

// GetScriptSource of a script by its scriptID.
func (t *Tab) GetScriptSource(scriptID string) (string, error) {
	scriptSrc, wasmSource, err := t.t.Debugger.GetScriptSource(scriptID)
	if wasmSource != "" {
		return wasmSource, err
	}
	return scriptSrc, err
}

// Gets the top document and updates our list of elements DO NOT CALL DOM.GetDocument after
// the page has loaded, it creates new nodeIDs and all functions that look up elements (QuerySelector)
// will fail.
func (t *Tab) getDocument() (*Element, error) {
	doc, err := t.t.DOM.GetDocument(-1, false)
	if err != nil {
		return nil, err
	}

	t.setTopNodeID(doc.NodeId)
	t.setTopFrameID(doc.FrameId)

	t.addNodes(doc)
	eleDoc, _ := t.getElement(doc.NodeId)
	return eleDoc, nil
}

// GetDocument returns the top level document element for this tab.
func (t *Tab) GetDocument() (*Element, error) {
	docEle, ok := t.getElement(t.GetTopNodeID())
	if !ok {
		return nil, &ElementNotFoundErr{Message: "top document node ID not found."}
	}
	return docEle, nil
}

// GetElementByNodeID returns either an element from our list of ready/known nodeIDs or a new un-ready element
// If it's not ready we return false. Note this does have a sIDe effect of adding a potentially
// invalID element to our list of known elements. But it is assumed this method will be called
// with a valID nodeID that chrome has not informed us about yet. Once we are informed, we need
// to update it via our list and not some reference that could disappear.
func (t *Tab) GetElementByNodeID(nodeID int) (*Element, bool) {
	t.eleMutex.RLock()
	ele, ok := t.elements[nodeID]
	t.eleMutex.RUnlock()
	if ok {
		return ele, true
	}
	newEle := newElement(t, nodeID)
	t.eleMutex.Lock()
	t.elements[nodeID] = newEle // add non-ready element to our list.
	t.eleMutex.Unlock()
	return newEle, false
}

// GetElementByLocation returns the element given the x, y coordinates on the page, or returns error.
func (t *Tab) GetElementByLocation(x, y int) (*Element, error) {
	_, _, nodeID, err := t.t.DOM.GetNodeForLocation(x, y, false, false)
	if err != nil {
		return nil, err
	}
	ele, _ := t.GetElementByNodeID(nodeID)
	return ele, nil
}

// GetAllElements returns a copy of all currently known elements. Note that modifications to elements
// maybe unsafe.
func (t *Tab) GetAllElements() map[int]*Element {
	t.eleMutex.RLock()
	allElements := make(map[int]*Element, len(t.elements))
	for k, v := range t.elements {
		allElements[k] = v
	}
	t.eleMutex.RUnlock()
	return allElements
}

// GetElementByID returns the element by searching the top level document for an element with attributeID
// Does not work on frames.
func (t *Tab) GetElementByID(attributeID string) (*Element, bool, error) {
	return t.GetDocumentElementByID(t.GetTopNodeID(), attributeID)
}

// GetDocumentElementByID returns an element from a specific Document.
func (t *Tab) GetDocumentElementByID(docNodeID int, attributeID string) (*Element, bool, error) {
	var err error

	docNode, ok := t.getElement(docNodeID)
	if !ok {
		return nil, false, &ElementNotFoundErr{Message: fmt.Sprintf("docNodeID %d not found", docNodeID)}
	}

	selector := "#" + attributeID

	nodeID, err := t.t.DOM.QuerySelector(docNode.ID, selector)
	if err != nil {
		return nil, false, err
	}
	ele, ready := t.GetElementByNodeID(nodeID)
	return ele, ready, nil
}

// GetElementsBySelector all elements that match a selector from the top level document
func (t *Tab) GetElementsBySelector(selector string) ([]*Element, error) {
	return t.GetDocumentElementsBySelector(t.GetTopNodeID(), selector)
}

// GetChildElements all elements of a child
func (t *Tab) GetChildElements(element *Element) []*Element {
	return t.GetChildElementsOfType(element, "*")
}

// GetChildElementsOfType all elements of a specific tag type.
func (t *Tab) GetChildElementsOfType(element *Element, tagType string) []*Element {
	elements := make([]*Element, 0)
	if element == nil || element.node == nil || element.node.Children == nil {
		return elements
	}
	t.recursivelyGetChildren(element.node.Children, &elements, tagType)
	return elements
}

// GetChildrensCharacterData the #text values of the element's children.
func (t *Tab) GetChildrensCharacterData(element *Element) string {
	var buf bytes.Buffer
	for _, el := range t.GetChildElements(element) {
		if el.nodeType == int(NodeText) {
			buf.WriteString(el.characterData)
		}
	}
	return buf.String()
}

func (t *Tab) recursivelyGetChildren(children []*gcdapi.DOMNode, elements *[]*Element, tagType string) {
	for _, child := range children {
		ele, ready := t.GetElementByNodeID(child.NodeId)
		// only add if it's ready and tagType matches or tagType is *
		if ready == true && (tagType == "*" || tagType == ele.nodeName) {
			*elements = append(*elements, ele)
		}
		// not ready, or doesn't have children
		if ready == false || ele.node.Children == nil || len(ele.node.Children) == 0 {
			continue
		}
		t.recursivelyGetChildren(ele.node.Children, elements, tagType)
	}
}

// GetDocumentElementsBySelector same as GetChildElementsBySelector
func (t *Tab) GetDocumentElementsBySelector(docNodeID int, selector string) ([]*Element, error) {
	docNode, ok := t.getElement(docNodeID)
	if !ok {
		return nil, &ElementNotFoundErr{Message: fmt.Sprintf("docNodeID %d not found", docNodeID)}
	}
	nodeIDs, errQuery := t.t.DOM.QuerySelectorAll(docNode.ID, selector)
	if errQuery != nil {
		return nil, errQuery
	}

	elements := make([]*Element, len(nodeIDs))

	for k, nodeID := range nodeIDs {
		elements[k], _ = t.GetElementByNodeID(nodeID)
	}

	return elements, nil
}

// GetElementsBySearch all elements that match a CSS or XPath selector
func (t *Tab) GetElementsBySearch(selector string, includeUserAgentShadowDOM bool) ([]*Element, error) {
	var s gcdapi.DOMPerformSearchParams
	s.Query = selector
	s.IncludeUserAgentShadowDOM = includeUserAgentShadowDOM
	ID, count, err := t.t.DOM.PerformSearchWithParams(&s)
	if err != nil {
		return nil, err
	}

	if count < 1 {
		return make([]*Element, 0), nil
	}

	var r gcdapi.DOMGetSearchResultsParams
	r.SearchId = ID
	r.FromIndex = 0
	r.ToIndex = count
	nodeIDs, errQuery := t.t.DOM.GetSearchResultsWithParams(&r)
	if errQuery != nil {
		return nil, errQuery
	}

	elements := make([]*Element, len(nodeIDs))

	for k, nodeID := range nodeIDs {
		elements[k], _ = t.GetElementByNodeID(nodeID)
	}

	return elements, nil
}

// GetPageSource returns the document's source, as visible, if docID is 0, returns top document source.
func (t *Tab) GetPageSource(docNodeID int) (string, error) {
	if docNodeID == 0 {
		docNodeID = t.GetTopNodeID()
	}
	doc, ok := t.getElement(docNodeID)
	if !ok {
		return "", &ElementNotFoundErr{Message: fmt.Sprintf("docNodeID %d not found", docNodeID)}
	}
	outerParams := &gcdapi.DOMGetOuterHTMLParams{NodeId: doc.ID}
	return t.t.DOM.GetOuterHTMLWithParams(outerParams)
}

// GetCurrentURL returns the current url of the top level document
func (t *Tab) GetCurrentURL() (string, error) {
	return t.GetDocumentCurrentURL(t.GetTopNodeID())
}

// GetDocumentCurrentURL returns the current url of the provIDed docNodeID
func (t *Tab) GetDocumentCurrentURL(docNodeID int) (string, error) {
	docNode, ok := t.getElement(docNodeID)
	if !ok {
		return "", &ElementNotFoundErr{Message: fmt.Sprintf("docNodeID %d not found", docNodeID)}
	}
	return docNode.node.DocumentURL, nil
}

// Screenshot returns a png image, base64 encoded, or error if failed
func (t *Tab) Screenshot(ctx context.Context) (string, error) {
	params := &gcdapi.PageCaptureScreenshotParams{
		Format:  "png",
		Quality: 100,
		Clip: &gcdapi.PageViewport{
			X:      0,
			Y:      0,
			Width:  1024,
			Height: 768,
			Scale:  float64(1)},
		FromSurface: true,
	}

	return t.t.Page.CaptureScreenshotWithParams(params)
}

// SerializeDOM and return it as string
func (t *Tab) SerializeDOM() string {
	node, err := t.t.DOM.GetDocument(-1, true)
	if err != nil {
		return ""
	}
	html, err := t.t.DOM.GetOuterHTMLWithParams(&gcdapi.DOMGetOuterHTMLParams{
		NodeId: node.NodeId,
	})
	if err != nil {
		return ""
	}
	return html
}

// GetNetworkTraffic returns all responses after page load
func (t *Tab) GetNetworkTraffic() (*Response, []*Response) {
	return t.container.GetResponses()
}

// CaptureNetworkTraffic ensures we capture all traffic (only saving text bodies) during navigation.
func (t *Tab) CaptureNetworkTraffic(ctx context.Context, URL string) {

	t.t.Network.EnableWithParams(&gcdapi.NetworkEnableParams{
		MaxPostDataSize:       -1,
		MaxResourceBufferSize: -1,
		MaxTotalBufferSize:    -1,
	})

	t.t.Subscribe("network.loadingFailed", func(target *gcd.ChromeTarget, payload []byte) {
		log.Info().Msgf("failed: %s\n", string(payload))
	})

	t.t.Subscribe("Network.requestWillBeSent", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.NetworkRequestWillBeSentEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		//message.Params.RedirectResponse.RemoteIPAddress
		if message.Params.Type == "Document" {
			//log.Info().Msgf("%s", string(payload))
			t.container.SetLoadRequest(message.Params.RequestId)
		}
	})

	t.t.Subscribe("Network.responseReceived", func(target *gcd.ChromeTarget, payload []byte) {
		//log.Info().Msgf("RESPONSE DATA: %#v\n", string(payload))
		defer t.container.DecRequest()
		t.container.IncRequest()

		message := &gcdapi.NetworkResponseReceivedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		p := message.Params

		//log.Ctx(ctx).Info().Str("request_ID", p.RequestID).Str("url", url).Msg("waiting")
		if err := t.container.WaitFor(timeoutCtx, p.RequestId); err != nil {
			return
		}

		// ignore file/data urls
		if strings.HasPrefix(p.Response.Url, "data") || strings.HasPrefix(p.Response.Url, "file") {
			return
		}

		t.buildResponse(URL, message)
		//log.Ctx(ctx).Info().Str("request_ID", p.RequestID).Str("url", url).Msg("adding response")
		//t.container.Add(response)
	})

	t.t.Subscribe("Network.loadingFinished", func(target *gcd.ChromeTarget, payload []byte) {
		//log.Info().Msgf("loadingFinished DATA: %#v\n", string(payload))
		message := &gcdapi.NetworkLoadingFinishedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		//log.Ctx(ctx).Info().Str("request_ID", message.Params.RequestID).Msg("finished")
		t.container.BodyReady(message.Params.RequestId)
	})
}

// Sets the element as invalid and removes it from our elements map
func (t *Tab) invalidateRemove(ele *Element) {
	ele.setInvalidated(true)
	t.eleMutex.Lock()
	delete(t.elements, ele.ID)
	t.eleMutex.Unlock()
}

// the entire document has been invalidated, request all nodes again
func (t *Tab) documentUpdated() {
	t.getDocument()
}

// Ask the debugger service for child nodes.
func (t *Tab) requestChildNodes(nodeId, depth int) {
	_, err := t.t.DOM.RequestChildNodes(nodeId, depth, false)
	if err != nil {
		log.Debug().Msgf("error requesting child nodes: %s\n", err)
	}
}

// Called if the element is known about but not yet populated. If it is not
// known, we create a new element. If it is known we populate it and return it.
func (t *Tab) nodeToElement(node *gcdapi.DOMNode) *Element {
	if ele, ok := t.getElement(node.NodeId); ok {
		ele.populateElement(node)
		return ele
	}
	newEle := newReadyElement(t, node)
	return newEle
}

// safely returns the element by looking it up by nodeId from our internal map.
func (t *Tab) getElement(nodeID int) (*Element, bool) {
	t.eleMutex.RLock()
	defer t.eleMutex.RUnlock()
	ele, ok := t.elements[nodeID]
	return ele, ok
}

// Safely adds the nodes in the document to our list of elements
// iterates over children and contentdocuments (if they exist)
// Calls requestchild nodes for each node so we can receive setChildNode
// events for even more nodes
func (t *Tab) addNodes(node *gcdapi.DOMNode) {
	newEle := t.nodeToElement(node)

	t.eleMutex.Lock()
	t.elements[newEle.ID] = newEle
	t.eleMutex.Unlock()
	//log.Printf("Added new element: %s\n", newEle)
	t.requestChildNodes(newEle.ID, 1)
	if node.Children != nil {
		// add child nodes
		for _, v := range node.Children {
			t.addNodes(v)
		}
	}
	if node.ContentDocument != nil {
		t.addNodes(node.ContentDocument)
	}
	t.lastNodeChangeTimeVal.Store(time.Now())
}

// buildResponse fills out a new with all relevant details
func (t *Tab) buildResponse(URL string, message *gcdapi.NetworkResponseReceivedEvent) {
	p := message.Params

	// set additional properties of web certificate

	if p.Type == "Document" {
		log.Info().Msg("is document")
	}

}

// encode the header depending on type, and lower case the header name so easier to search in DB.
func (t *Tab) encodeHeaders(gcdHeaders map[string]interface{}) map[string]string {
	headers := make(map[string]string, len(gcdHeaders))
	for k, v := range gcdHeaders {
		name := strings.ToLower(k)
		switch rv := v.(type) {
		case string:
			headers[name] = rv
		case []string:
			headers[name] = strings.Join(rv, ",")
		case nil:
			headers[name] = ""
		default:
			log.Warn().Str("header_name", k).Msg("unable to encode header value")
		}
	}
	return headers
}

func (t *Tab) extractCertificate(message *gcdapi.NetworkResponseReceivedEvent) error {
	p := message.Params

	u, err := url.Parse(p.Response.Url)
	log.Info().Msgf("url: %s\n", u)

	/*
		if u.Hostname() == t.address && u.Scheme == "https" &&
			strings.HasPrefix(p.Response.Url, "https") && p.Response.SecurityDetails != nil {
				cert := convert.NetworkCertificateToWebCertificate(p.Response.SecurityDetails)
				cert.AddressHash = convert.HashAddress(ipAddress, host)
				cert.IPAddress = ipAddress
				cert.Port = port
		}
	*/
	return err
}

func (t *Tab) encodeResponseBody(p *gcdapi.NetworkResponseReceivedEvent) string {

	var err error
	var encoded bool
	var body []byte
	var bodyStr string

	bodyStr, encoded, err = t.t.Network.GetResponseBody(p.Params.RequestId)
	if err != nil {
		log.Warn().Str("url", p.Params.Response.Url).Err(err).Msg("failed to get body")
	}

	body = []byte(bodyStr)
	if encoded {
		body, _ = base64.StdEncoding.DecodeString(bodyStr)
	}

	// we don't want to capture anything other than text based files.
	if !strings.HasPrefix(http.DetectContentType(body), "text") {
		bodyStr = ""
	}

	return bodyStr
}

func (t *Tab) domUpdated(ctx context.Context) func(target *gcd.ChromeTarget, payload []byte) {
	return func(target *gcd.ChromeTarget, payload []byte) {
		//log.Ctx(ctx).Info().Msg("dom updated")
		t.lastNodeChangeTimeVal.Store(time.Now())
	}
}

func (t *Tab) subscribeBrowserEvents(ctx context.Context) {
	t.t.DOM.Enable()
	t.t.Inspector.Enable()
	t.t.Page.Enable()
	t.t.Security.Enable()
	t.t.Console.Enable()
	t.t.Debugger.Enable(-1)

	t.t.Security.SetOverrideCertificateErrors(true)

	t.t.Subscribe("Security.certificateError", func(target *gcd.ChromeTarget, payload []byte) {
		resp := &gcdapi.SecurityCertificateErrorEvent{}
		err := json.Unmarshal(payload, resp)
		if err != nil {
			return
		}
		log.Info().Str("type", resp.Params.ErrorType).Msg("handling certificate error")
		p := &gcdapi.SecurityHandleCertificateErrorParams{
			EventId: resp.Params.EventId,
			Action:  "continue",
		}

		t.t.Security.HandleCertificateErrorWithParams(p)
		log.Ctx(ctx).Info().Msg("certificate error handled")
	})

	t.t.Subscribe("Inspector.targetCrashed", func(target *gcd.ChromeTarget, payload []byte) {
		log.Ctx(ctx).Warn().Msgf("tab crashed: %s", string(payload))
		select {
		case t.crashedCh <- "crashed":
		case <-t.exitCh:
		}
	})

	t.t.Subscribe("Inspector.detached", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.InspectorDetachedEvent{}
		err := json.Unmarshal(payload, header)
		reason := "detached"

		if err == nil {
			reason = header.Params.Reason
		}

		select {
		case t.crashedCh <- reason:
		case <-t.exitCh:
		}
	})

	t.t.Subscribe("Page.loadEventFired", func(target *gcd.ChromeTarget, payload []byte) {
		select {
		case t.navigationCh <- 0:
		case <-t.exitCh:
		}
	})

	// new nodes
	t.t.Subscribe("DOM.setChildNodes", t.domUpdated(ctx))
	t.t.Subscribe("DOM.attributeModified", t.domUpdated(ctx))
	t.t.Subscribe("DOM.attributeRemoved", t.domUpdated(ctx))
	t.t.Subscribe("DOM.characterDataModified", t.domUpdated(ctx))
	t.t.Subscribe("DOM.childNodeCountUpdated", t.domUpdated(ctx))
	t.t.Subscribe("DOM.childNodeInserted", t.domUpdated(ctx))
	t.t.Subscribe("DOM.childNodeRemoved", t.domUpdated(ctx))
	t.t.Subscribe("DOM.documentUpdated", t.domUpdated(ctx))

}
