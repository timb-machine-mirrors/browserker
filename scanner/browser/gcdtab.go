package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"

	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

// Tab is a chromium browser tab we use for instrumentation
type Tab struct {
	g         *gcd.Gcd
	t         *gcd.ChromeTarget
	ctx       *browserk.Context
	container *Container
	id        int64
	eleMutex  *sync.RWMutex           // locks our elements when added/removed.
	elements  map[int]*gcdapi.DOMNode // our map of elements for this tab

	topNodeID             atomic.Value           // the nodeID of the current top level #document
	topFrameID            atomic.Value           // the frameID of the current top level #document
	baseHref              atomic.Value           // the base href for the current top document
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
	stableAfter           time.Duration          // amount of time of no activity to consider the DOM stable
	lastNodeChangeTimeVal atomic.Value           // timestamp of when the last node change occurred atomic because multiple go routines will modify
	domChangeHandler      DomChangeHandlerFunc   // allows the caller to be notified of DOM change events.
	docWasUpdated         atomic.Value           // for tracking if an execution caused a new page load/transition

	frameMutex *sync.RWMutex
	frames     map[string]int // frames
}

// NewTab to use
func NewTab(ctx *browserk.Context, gcdBrowser *gcd.Gcd, tab *gcd.ChromeTarget) *Tab {
	id := rand.Int63() // TODO: generate random or something
	t := &Tab{
		t:            tab,
		ctx:          ctx,
		container:    NewContainer(),
		crashedCh:    make(chan string),
		exitCh:       make(chan struct{}),
		navigationCh: make(chan int),
	}
	t.id = id
	t.g = gcdBrowser
	t.eleMutex = &sync.RWMutex{}
	t.elements = make(map[int]*gcdapi.DOMNode)

	t.frames = make(map[string]int)
	t.frameMutex = &sync.RWMutex{}

	t.nodeChange = make(chan *NodeChangeEvent)
	t.navigationCh = make(chan int, 1)  // for signaling navigation complete
	t.docUpdateCh = make(chan struct{}) // wait for documentUpdate to be called during navigation
	t.crashedCh = make(chan string)     // reason the tab crashed/was disconnected.
	t.exitCh = make(chan struct{})
	t.navigationTimeout = 30 * time.Second // default 30 seconds for timeout
	t.elementTimeout = 5 * time.Second     // default 5 seconds for waiting for element.
	t.stabilityTimeout = 2 * time.Second   // default 2 seconds before we give up waiting for stability
	t.stableAfter = 300 * time.Millisecond // default 300 ms for considering the DOM stable
	t.domChangeHandler = nil

	t.disconnectedHandler = t.defaultDisconnectedHandler
	go t.listenDebuggerEvents(ctx)
	t.subscribeBrowserEvents(ctx, true)
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

// ExecuteAction for this browser, calling js handler after it is called
func (t *Tab) ExecuteAction(ctx context.Context, act *browserk.Action) ([]byte, bool, error) {
	var err error
	causedLoad := false
	// Call JSBefore hooks
	t.ctx.NextJSBefore(t)

	// reset doc was updated flag
	t.docWasUpdated.Store(false)
	// do action
	switch act.Type {
	case browserk.ActLoadURL:
		t.Navigate(ctx, string(act.Input))
	case browserk.ActExecuteJS:
		t.InjectJS(string(act.Input))
	case browserk.ActLeftClick:
		ele, err := t.FindByHTMLElement(act.Element)
		if err != nil {
			return nil, false, err
		}
		err = ele.Click()
		// add small delay after click
		timer := time.NewTimer(time.Millisecond * 200)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
		}

		if t.IsTransitioning() {
			log.Info().Msg("CLICK CAUSED TRANSITION EVENT!")
			t.waitReady(ctx, t.stabilityTimeout)
		}

	case browserk.ActLeftClickDown:
	case browserk.ActLeftClickUp:
	case browserk.ActRightClick:
	case browserk.ActRightClickDown:
	case browserk.ActRightClickUp:
	case browserk.ActMiddleClick:
	case browserk.ActMiddleClickDown:
	case browserk.ActMiddleClickUp:
	case browserk.ActScroll:
	case browserk.ActSendKeys:
	case browserk.ActKeyUp:
	case browserk.ActKeyDown:
	case browserk.ActHover:
	case browserk.ActFocus:
	case browserk.ActWait:
	}
	// Call JSAfter hooks

	t.ctx.NextJSAfter(t)
	if docUpdated, ok := t.docWasUpdated.Load().(bool); ok {
		causedLoad = docUpdated
	}

	return nil, causedLoad, err
}

// Navigate to the url
func (t *Tab) Navigate(ctx context.Context, url string) error {
	if t.IsNavigating() {
		return &InvalidNavigationErr{Message: "Unable to navigate, already navigating."}
	}

	t.setIsNavigating(true)
	defer t.setIsNavigating(false)
	log.Debug().Msgf("navigating to %s", url)
	navParams := &gcdapi.PageNavigateParams{Url: url, TransitionType: "typed"}
	frameID, _, errText, err := t.t.Page.NavigateWithParams(navParams)
	if err != nil {
		return err
	}
	t.setTopFrameID(frameID)

	if errText != "" {
		return errors.Wrap(ErrNavigating, errText)
	}

	t.lastNodeChangeTimeVal.Store(time.Now())
	log.Debug().Msgf("waiting ready for %s", url)
	return t.waitReady(ctx, t.stableAfter)
}

// IsShuttingDown answers if we are shutting down or not
func (t *Tab) IsShuttingDown() bool {
	if flag, ok := t.shutdown.Load().(bool); ok {
		return flag
	}
	return false
}

func (t *Tab) setShutdownState(val bool) {
	t.shutdown.Store(val)
}

// ID of this browser (tab)
func (t *Tab) ID() int64 {
	return t.id
}

// FindByHTMLElement returns a gcd Element for interacting
func (t *Tab) FindByHTMLElement(ele *browserk.HTMLElement) (*Element, error) {
	if ele == nil {
		return nil, &ErrInvalidElement{}
	}
	tag := strings.ToLower(browserk.HTMLTypeToStrMap[ele.Type])
	if ele.Type == browserk.CUSTOM {
		tag = ele.CustomTagName
	}

	eles, err := t.GetElementsBySelector(tag)
	if err != nil {
		return nil, err
	}

	for _, found := range eles {
		h := ElementToHTMLElement(found)
		if bytes.Compare(h.Hash(), ele.Hash()) == 0 && h.Depth == ele.Depth {
			log.Info().Msg("found by nearly exact match")
			return found, nil
		}
	}
	return nil, &ElementNotFoundErr{}
}

// FindElements elements via querySelector, does not pull out children
func (t *Tab) FindElements(querySelector string) ([]*browserk.HTMLElement, error) {
	elements, err := t.GetElementsBySelector(querySelector)
	if err != nil {
		return nil, err
	}

	bElements := make([]*browserk.HTMLElement, 0)
	for _, ele := range elements {
		bElements = append(bElements, ElementToHTMLElement(ele))
	}
	return bElements, nil
}

// GetBaseHref of the top level document
// TODO will need to handle iframes here too
func (t *Tab) GetBaseHref() string {
	return t.baseHref.Load().(string)
}

// FindForms finds forms and pulls out all child elements.
// we may need more than just input fields (labels) etc for context
func (t *Tab) FindForms() ([]*browserk.HTMLFormElement, error) {
	elements, err := t.GetElementsBySelector("form")
	if err != nil {
		return nil, err
	}
	fElements := make([]*browserk.HTMLFormElement, 0)
	for _, form := range elements {
		f := &browserk.HTMLFormElement{
			Events:        make([]browserk.HTMLEventType, 0),
			ChildElements: make([]*browserk.HTMLElement, 0),
		}

		childNodes, _ := form.GetChildNodeIds()
		for _, childID := range childNodes {
			child, _ := t.getElementByNodeID(childID)
			child.WaitForReady()
			f.ChildElements = append(f.ChildElements, ElementToHTMLElement(child))
		}
		fElements = append(fElements, f)
	}
	return fElements, nil
}

// GetMessages that occurred since last called
func (t *Tab) GetMessages() ([]*browserk.HTTPMessage, error) {
	msgs := t.container.GetMessages()
	return msgs, nil
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

// GetNavURL by looking at the navigation history
func (t *Tab) GetNavURL() string {
	_, entries, err := t.t.Page.GetNavigationHistory()
	if err != nil || len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Url
}

// WaitReady waits for the page to load, DOM to be stable, and no network traffic in progress
func (t *Tab) waitReady(ctx context.Context, stableAfter time.Duration) error {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	navTimer := time.After(45 * time.Second)
	// wait navigation to complete.
	log.Info().Msg("waiting for nav to complete")
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
	defer t.t.Page.GetResourceTree()
	// wait for DOM & network stability
	log.Info().Msg("waiting for nav stability complete")
	for {
		select {
		case reason := <-t.crashedCh:
			return errors.Wrap(ErrTabCrashed, reason)
		case <-ctx.Done():
			return ctx.Err()
		case <-t.exitCh:
			return ErrTabClosing
		case <-stableTimer:
			log.Info().Msg("stability timed out")
			return ErrTimedOut
		case <-ticker.C:
			if changeTime, ok := t.lastNodeChangeTimeVal.Load().(time.Time); ok {
				log.Info().Int32("requests", t.container.OpenRequestCount()).Msgf("tick %s >= %s", time.Now().Sub(changeTime), stableAfter)
				if time.Now().Sub(changeTime) >= stableAfter && t.container.OpenRequestCount() == 0 {
					// times up, should be stable now
					log.Info().Msg("stable")
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

// SetStabilityTime to wait for no node changes before we consider the DOM stable.
// Note that stability timeout will fire if the DOM is constantly changing.
// The deafult stableAfter is 300 ms.
func (t *Tab) SetStabilityTime(stableAfter time.Duration) {
	t.stableAfter = stableAfter
}

func (t *Tab) setIsNavigating(set bool) {
	t.isNavigatingFlag.Store(set)
	t.baseHref.Store("")
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
	t.baseHref.Store("")
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

// getTopFrameID return the top frame ID of this tab
func (t *Tab) getTopFrameID() string {
	if frameID, ok := t.topFrameID.Load().(string); ok {
		return frameID
	}
	return ""
}

func (t *Tab) setTopNodeID(nodeID int) {
	t.topNodeID.Store(nodeID)
}

// getTopNodeID returns the current top node ID of this Tab.
func (t *Tab) getTopNodeID() int {
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

// GetCookies from the browser
func (t *Tab) GetCookies() ([]*browserk.Cookie, error) {
	cookies, err := t.t.Page.GetCookies()
	if err != nil {
		return nil, err
	}
	return GCDCookieToBrowserk(cookies), nil
}

// GetStorageEvents and clear the container
func (t *Tab) GetStorageEvents() []*browserk.StorageEvent {
	return t.container.GetStorageEvents()
}

// GetConsoleEvents and clear the container
func (t *Tab) GetConsoleEvents() []*browserk.ConsoleEvent {
	return t.container.GetConsoleEvents()
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

// Gets the top document and updates our list of elements it creates all new nodeIDs.
func (t *Tab) getDocument() (*Element, error) {
	log.Debug().Msgf("getDocument doc id was: %d", t.getTopNodeID())
	doc, err := t.t.DOM.GetDocument(-1, true)
	if err != nil {
		return nil, err
	}
	t.setTopNodeID(doc.NodeId)
	log.Debug().Msgf("getDocument doc id is now: %d", t.getTopNodeID())
	t.addNodes(doc, 0)
	eleDoc, _ := t.getElement(doc.NodeId)
	return eleDoc, nil
}

// GetDocument returns the top level document element for this tab.
func (t *Tab) GetDocument() (*Element, error) {
	docEle, ok := t.getElement(t.getTopNodeID())
	if !ok {
		return nil, &ElementNotFoundErr{Message: "top document node ID not found."}
	}
	return docEle, nil
}

// getElementByNodeID returns either an element from our list of ready/known nodeIDs or a new un-ready element
// If it's not ready we return false. Note this does have a side effect of adding a potentially
// invalid element to our list of known elements. But it is assumed this method will be called
// with a valid nodeID that chrome has not informed us about yet. Once we are informed, we need
// to update it via our list and not some reference that could disappear.
func (t *Tab) getElementByNodeID(nodeID int) (*gcdapi.DOMNode, bool) {
	t.eleMutex.RLock()
	node, ok := t.elements[nodeID]
	t.eleMutex.RUnlock()
	if ok {
		return node, true
	}
	node = &gcdapi.DOMNode{NodeId: nodeID}
	t.eleMutex.Lock()
	t.elements[nodeID] = node // add non-ready element to our list.
	t.eleMutex.Unlock()
	return node, false
}

// GetElementByLocation returns the element given the x, y coordinates on the page, or returns error.
func (t *Tab) GetElementByLocation(x, y int) (*Element, error) {
	_, _, nodeID, err := t.t.DOM.GetNodeForLocation(x, y, false, false)
	if err != nil {
		return nil, err
	}
	ele, _ := t.getElementByNodeID(nodeID)
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
	return t.getDocumentElementByID(t.getTopNodeID(), attributeID)
}

// getDocumentElementByID returns an element from a specific Document.
func (t *Tab) getDocumentElementByID(docNodeID int, attributeID string) (*Element, bool, error) {
	var err error

	selector := "#" + attributeID

	nodeID, err := t.t.DOM.QuerySelector(docNodeID, selector)
	if err != nil {
		return nil, false, err
	}
	ele, ready := t.getElementByNodeID(nodeID)
	return ele, ready, nil
}

// GetElementsBySelector all elements that match a selector from the top level document
// also searches sub frames
func (t *Tab) GetElementsBySelector(selector string) ([]*Element, error) {
	elements, err := t.GetDocumentElementsBySelector(t.getTopNodeID(), selector)
	if err != nil {
		// try again but refresh the doc
		t.RefreshDocument()
		elements, err = t.GetDocumentElementsBySelector(t.getTopNodeID(), selector)
		if err != nil {
			return nil, err
		}
	}

	// search frames too
	frameNodeIDs := t.getFrameNodeIDs()
	for _, id := range frameNodeIDs {
		frameElements, err := t.GetDocumentElementsBySelector(id, selector)
		if err != nil {
			log.Warn().Msg("failed to search frame for elements")
			continue
		}
		elements = append(elements, frameElements...)
	}
	return elements, err
}

func (t *Tab) getFrameNodeIDs() []int {
	nodeIDs := make([]int, 0)
	t.frameMutex.RLock()
	for _, v := range t.frames {
		nodeIDs = append(nodeIDs, v)
	}
	t.frameMutex.RUnlock()
	return nodeIDs
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
		ele, ready := t.getElementByNodeID(child.NodeId)
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
	nodeIDs, errQuery := t.t.DOM.QuerySelectorAll(docNodeID, selector)
	if errQuery != nil {
		log.Info().Msgf("QuerySelectorAll Err: searching for %s %d", selector, docNodeID)
		return nil, errQuery
	}

	elements := make([]*Element, len(nodeIDs))

	for k, nodeID := range nodeIDs {
		elements[k], _ = t.getElementByNodeID(nodeID)
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
		elements[k], _ = t.getElementByNodeID(nodeID)
	}

	return elements, nil
}

// GetDOM in serialized form
func (t *Tab) GetDOM() (string, error) {
	node, err := t.t.DOM.GetDocument(-1, true)
	if err != nil {
		return "", err
	}
	html, err := t.t.DOM.GetOuterHTMLWithParams(&gcdapi.DOMGetOuterHTMLParams{
		NodeId: node.NodeId,
	})
	return html, err
}

// GetPageSource returns the document's source, as visible, if docID is 0, returns top document source.
func (t *Tab) GetPageSource(docNodeID int) (string, error) {
	if docNodeID == 0 {
		docNodeID = t.getTopNodeID()
	}
	doc, ok := t.getElement(docNodeID)
	if !ok {
		return "", &ElementNotFoundErr{Message: fmt.Sprintf("docNodeID %d not found", docNodeID)}
	}
	outerParams := &gcdapi.DOMGetOuterHTMLParams{NodeId: doc.ID}
	return t.t.DOM.GetOuterHTMLWithParams(outerParams)
}

// GetURL returns the current url of the top level document
func (t *Tab) GetURL() (string, error) {
	return t.GetDocumentCurrentURL(t.getTopNodeID())
}

// GetDocumentCurrentURL returns the current url of the provIDed docNodeID
func (t *Tab) GetDocumentCurrentURL(docNodeID int) (string, error) {
	docNode, ok := t.getElement(docNodeID)
	if !ok {
		return "", &ElementNotFoundErr{Message: fmt.Sprintf("docNodeID %d not found", docNodeID)}
	}
	return docNode.DocumentURL, nil
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

// Sets the element as invalid and removes it from our elements map
func (t *Tab) invalidateRemove(ele *Element) {
	ele.setInvalidated(true)
	t.eleMutex.Lock()
	delete(t.elements, ele.ID)
	t.eleMutex.Unlock()
}

// the entire document has been invalidated, request all nodes again
func (t *Tab) documentUpdated() {
	log.Info().Msg("document has been invalidated")
	t.docWasUpdated.Store(true)
	t.getDocument()
}

// Ask the debugger service for child nodes.
func (t *Tab) requestChildNodes(nodeID, depth int) {
	_, err := t.t.DOM.RequestChildNodes(nodeID, depth, false)
	if err != nil {
		log.Debug().Msgf("error requesting child nodes: %s\n", err)
	}
}

// safely returns the element by looking it up by nodeId from our internal map.
func (t *Tab) getNode(nodeID int) (*gcdapi.DOMNode, bool) {
	t.eleMutex.RLock()
	defer t.eleMutex.RUnlock()
	ele, ok := t.elements[nodeID]
	return ele, ok
}

// Safely adds the nodes in the document to our list of elements
// iterates over children and contentdocuments (if they exist)
// Calls requestchild nodes for each node so we can receive setChildNode
// events for even more nodes
func (t *Tab) addNodes(node *gcdapi.DOMNode, depth int) {
	t.eleMutex.Lock()
	t.elements[node.NodeId] = node
	t.eleMutex.Unlock()

	if node.Children != nil {
		// add child nodes
		for _, v := range node.Children {
			t.addNodes(v, depth+1)
		}
	}

	// base href can cause relative links to go out of scope
	// so we need to capture it
	tag := node.NodeName
	if tag == "BASE" && NodeHasAttribute(node, "href") {
		t.baseHref.Store(NodeGetAttribute(node, "href"))
	}

	if node.ContentDocument != nil {
		t.frameMutex.Lock()
		t.frames[node.FrameId] = node.ContentDocument.NodeId
		t.frameMutex.Unlock()

		t.addNodes(node.ContentDocument, depth+1)
	}
	t.lastNodeChangeTimeVal.Store(time.Now())
}

// Listens for NodeChangeEvents and crash events, dispatches them accordingly.
// Calls the user defined domChangeHandler if bound. Updates the lastNodeChangeTime
// to the current time. If the target crashes or is detached, call the disconnectedHandler.
func (t *Tab) listenDebuggerEvents(ctx *browserk.Context) {
	for {
		select {
		case nodeChangeEvent := <-t.nodeChange:
			t.lastNodeChangeTimeVal.Store(time.Now())
			t.handleNodeChange(nodeChangeEvent)
			// if the caller registered a dom change listener, call it
			if t.domChangeHandler != nil {
				t.domChangeHandler(t, nodeChangeEvent)
			}
		case reason := <-t.crashedCh:
			if t.disconnectedHandler != nil {
				go t.disconnectedHandler(t, reason)
			}
		case <-t.exitCh:
			log.Info().Msg("exiting...")
			return
		case <-ctx.Ctx.Done():
			log.Info().Msg("context done exiting...")
			return
		}
	}
}

func (t *Tab) RefreshDocument() {
	t.handleDocumentUpdated()
}

// Handles the document updated event. This occurs after a navigation or redirect.
// This is a destructive action which invalidates all document nodeids and their children.
// We loop through our current list of elements and invalidate them so any references
// can check if they are valid or not. We then recreate the elements map. Finally, if we
// are navigating, we want to block Navigate from returning until we have a valid document,
// so we use the docUpdateCh to signal when complete.
func (t *Tab) handleDocumentUpdated() {
	// set all elements as invalid and destroy the Elements map.
	t.eleMutex.Lock()
	t.elements = make(map[int]*gcdapi.DOMNode)
	t.eleMutex.Unlock()

	t.documentUpdated()
	// notify if navigating that we received the document update event.
	if t.IsNavigating() {
		t.docUpdateCh <- struct{}{} // notify listeners document was updated
	}
}

// handle node change events, updating, inserting invalidating and removing
func (t *Tab) handleNodeChange(change *NodeChangeEvent) {
	// if we are shutting down, do not handle new node changes.
	if t.IsShuttingDown() {
		return
	}

	switch change.EventType {
	case DocumentUpdatedEvent:
		t.handleDocumentUpdated()
	case SetChildNodesEvent:
		t.handleSetChildNodes(change.ParentNodeID, change.Nodes)
	case AttributeModifiedEvent:
		t.eleMutex.Lock()
		if node, ok := t.elements[change.NodeID]; ok {
			if NodeHasAttribute(node, change.Name) {
				NodeUpdateAttribute(node, change.Name, change.Value)
			}
		}
		t.eleMutex.Unlock()
	case AttributeRemovedEvent:
		t.eleMutex.Lock()
		if node, ok := t.elements[change.NodeID]; ok {
			if NodeHasAttribute(node, change.Name) {
				NodeRemoveAttribute(node, change.Name, change.Value)
			}
		}
		t.eleMutex.Unlock()
	case CharacterDataModifiedEvent:
		t.eleMutex.Lock()
		if node, ok := t.elements[change.NodeID]; ok {
			node.NodeValue = change.CharacterData
		}
		t.eleMutex.Unlock()
	case ChildNodeCountUpdatedEvent:
		t.eleMutex.Lock()
		if node, ok := t.elements[change.NodeID]; ok {
			node.ChildNodeCount = change.ChildNodeCount
			t.eleMutex.Unlock()
			// request the child nodes
			t.requestChildNodes(change.NodeID, 1)
		} else {
			t.eleMutex.Unlock()
		}

	case ChildNodeInsertedEvent:
		t.handleChildNodeInserted(change.ParentNodeID, change.Node)
	case ChildNodeRemovedEvent:
		t.handleChildNodeRemoved(change.ParentNodeID, change.NodeID)
	}

}

// setChildNode event handling will add nodes to our elements map and update
// the parent reference Children
func (t *Tab) handleSetChildNodes(parentNodeID int, nodes []*gcdapi.DOMNode) {
	parent, ok := t.getElementByNodeID(parentNodeID)
	depth := parent.Depth() + 1
	for _, node := range nodes {
		t.addNodes(node, depth)
	}
	if ok {
		if err := parent.WaitForReady(); err == nil {
			parent.addChildren(nodes)
		}
	}
	t.lastNodeChangeTimeVal.Store(time.Now())

}

// update parent with new child node and add the new nodes.
func (t *Tab) handleChildNodeInserted(parentNodeID int, node *gcdapi.DOMNode) {
	t.lastNodeChangeTimeVal.Store(time.Now())
	if node == nil {
		return
	}
	parent, _ := t.getElementByNodeID(parentNodeID)
	depth := parent.Depth() + 1
	t.addNodes(node, depth)

	// make sure we have the parent before we add children
	if err := parent.WaitForReady(); err == nil {
		parent.addChild(node)
		return
	}

}

// Update ParentNodeId to remove child and iterate over Children recursively and invalidate them.
// TODO: come up with a better way of removing children without direct access to the node
// as it's a potential race condition if it's being modified.
func (t *Tab) handleChildNodeRemoved(parentNodeID, nodeID int) {
	ele, ok := t.getElement(nodeID)
	if !ok {
		return
	}
	ele.setInvalidated(true)
	parent, ok := t.getElement(parentNodeID)

	if ok {
		if err := parent.WaitForReady(); err == nil {
			parent.removeChild(ele.NodeID())
		}
	}

	// if not ready, node will be nil
	if ele.IsReadyInvalid() {
		t.invalidateChildren(ele.node)
	}

	t.eleMutex.Lock()
	delete(t.elements, nodeID)
	t.eleMutex.Unlock()
}

// when a childNodeRemoved event occurs, we need to set each child
// to invalidated and remove it from our elements map.
func (t *Tab) invalidateChildren(node *gcdapi.DOMNode) {
	// invalidate & remove ContentDocument node and children
	if node.ContentDocument != nil {
		ele, ok := t.getElement(node.ContentDocument.NodeId)
		if ok {
			t.invalidateRemove(ele)
			t.invalidateChildren(node.ContentDocument)
		}
	}

	if node.Children == nil {
		return
	}

	// invalidate node.Children
	for _, child := range node.Children {
		ele, ok := t.getElement(child.NodeId)
		if !ok {
			continue
		}
		t.invalidateRemove(ele)
		// recurse and remove children of this node
		t.invalidateChildren(ele.node)
	}
}

func (t *Tab) subscribeBrowserEvents(ctx *browserk.Context, intercept bool) {
	t.t.DOM.Enable()
	t.t.Inspector.Enable()
	t.t.Page.Enable()
	t.t.Security.Enable()
	t.t.Console.Enable()
	t.t.Debugger.Enable(-1)

	t.t.Network.EnableWithParams(&gcdapi.NetworkEnableParams{
		MaxPostDataSize:       -1,
		MaxResourceBufferSize: -1,
		MaxTotalBufferSize:    -1,
	})

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
		log.Info().Msg("certificate error handled")
	})

	// network related events
	t.subscribeNetworkEvents(ctx)
	if intercept {
		patterns := []*gcdapi.FetchRequestPattern{
			{
				UrlPattern:   "*",
				RequestStage: "Request",
			},
			{
				UrlPattern:   "*",
				RequestStage: "Response",
			},
		}
		t.t.Fetch.EnableWithParams(&gcdapi.FetchEnableParams{
			Patterns:           patterns,
			HandleAuthRequests: false,
		})
		t.subscribeInterception(ctx)
	}
	// crash related events
	t.subscribeTargetCrashed()
	t.subscribeTargetDetached()

	// load releated events
	t.subscribeLoadEvent()
	t.subscribeFrameLoadingEvent()
	t.subscribeFrameFinishedEvent()

	// DOM update related events
	t.subscribeDocumentUpdated()
	t.subscribeSetChildNodes()
	t.subscribeAttributeModified()
	t.subscribeAttributeRemoved()
	t.subscribeCharacterDataModified()
	t.subscribeChildNodeCountUpdated()
	t.subscribeChildNodeInserted()
	t.subscribeChildNodeRemoved()

	// events
	t.subscribeStorageEvents()
	t.subscribeConsoleEvents()
}
