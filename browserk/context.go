package browserk

import "math"

// RequestHandler for adding middleware between browser HTTP Request events
type RequestHandler func(c *Context)

// ResponseHandler for adding middleware between browser HTTP Response events
type ResponseHandler func(c *Context)

// EventHandler for adding middleware between browser events
type EventHandler func(c *Context)

const abortIndex int8 = math.MaxInt8 / 2

// Context shared between services, browsers and plugins
type Context struct {
	Scope    ScopeService
	Reporter Reporter
	Injector Injector

	reqHandlers []RequestHandler
	reqIndex    int8

	respHandlers []ResponseHandler
	respIndex    int8

	evtHandlers []EventHandler
	evtIndex    int8
}

// NextReq calls the next handler
func (c *Context) NextReq() {
	c.reqIndex++
	for c.reqIndex < int8(len(c.reqHandlers)) {
		c.reqHandlers[c.reqIndex-1](c)
		c.reqIndex++
	}
}

// AddReqHandler adds new request handlers
func (c *Context) AddReqHandler(i ...RequestHandler) {
	if c.reqHandlers == nil {
		c.reqHandlers = make([]RequestHandler, 0)
	}
	c.reqHandlers = append(c.reqHandlers, i...)
}

// IsReqAborted returns true if the current context was aborted.
func (c *Context) IsReqAborted() bool {
	return c.reqIndex >= abortIndex
}

// ReqAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) ReqAbort() {
	c.reqIndex = abortIndex
}

// NextResp calls the next handler
func (c *Context) NextResp() {
	c.respIndex++
	for c.respIndex < int8(len(c.respHandlers)) {
		c.respHandlers[c.respIndex-1](c)
		c.respIndex++
	}

}

// AddRespHandler adds new request handlers
func (c *Context) AddRespHandler(i ...ResponseHandler) {
	if c.respHandlers == nil {
		c.respHandlers = make([]ResponseHandler, 0)
	}
	c.respHandlers = append(c.respHandlers, i...)
}

// IsRespAborted returns true if the current context was aborted.
func (c *Context) IsRespAborted() bool {
	return c.respIndex >= abortIndex
}

// RespAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) RespAbort() {
	c.respIndex = abortIndex
}

// NextEvt calls the next handler
func (c *Context) NextEvt() {
	c.evtIndex++
	for c.evtIndex < int8(len(c.evtHandlers)) {
		c.evtHandlers[c.evtIndex-1](c)
		c.evtIndex++
	}
}

// AddEvtHandler adds new request handlers
func (c *Context) AddEvtHandler(i ...EventHandler) {
	if c.evtHandlers == nil {
		c.evtHandlers = make([]EventHandler, 0)
	}
	c.evtHandlers = append(c.evtHandlers, i...)
}

// IsEvtAborted returns true if the current context was aborted.
func (c *Context) IsEvtAborted() bool {
	return c.evtIndex >= abortIndex
}

// EvtAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) EvtAbort() {
	c.evtIndex = abortIndex
}