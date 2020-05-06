package browser

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

type Container struct {
	messageLock   sync.RWMutex
	messages      map[string]*browserk.HTTPMessage
	loadRequestID string

	respReadyLock sync.RWMutex
	respReady     map[string]chan struct{}

	requestCount int32
}

func NewContainer() *Container {
	return &Container{
		messages:  make(map[string]*browserk.HTTPMessage),
		respReady: make(map[string]chan struct{}),
	}
}

// SetLoadRequest uses the requestID of the *first* request as
// our key to return the httpresponse in GetResponses.
func (c *Container) SetLoadRequest(request *browserk.HTTPRequest) {
	c.messageLock.Lock()
	defer c.messageLock.Unlock()
	if c.loadRequestID != "" {
		return
	}
	c.loadRequestID = request.RequestId
}

// GetMessages returns all of the http req/resp messages
func (c *Container) GetMessages() []*browserk.HTTPMessage {
	c.messageLock.Lock()
	defer c.messageLock.Unlock()
	msgs := make([]*browserk.HTTPMessage, len(c.messages))
	i := 0
	log.Info().Msgf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!we have %d messages", len(c.messages))
	for _, v := range c.messages {
		if v == nil {
			continue
		}
		msgs[i] = v
		i++
	}
	c.messages = make(map[string]*browserk.HTTPMessage, 0)
	c.loadRequestID = ""
	return msgs
}

// AddRequest to our map of messages
func (c *Container) AddRequest(request *browserk.HTTPRequest) {
	var exist bool
	msg := &browserk.HTTPMessage{}

	c.messageLock.Lock()
	if msg, exist = c.messages[request.RequestId]; !exist {
		msg = &browserk.HTTPMessage{}
	}
	msg.Request = request
	c.messages[request.RequestId] = msg
	c.messageLock.Unlock()
}

// AddResponse to our map of messages
func (c *Container) AddResponse(response *browserk.HTTPResponse) {
	var exist bool
	msg := &browserk.HTTPMessage{}
	log.Info().Str("request_id", response.RequestId).Msg("adding response")
	c.messageLock.Lock()
	if msg, exist = c.messages[response.RequestId]; !exist {
		msg = &browserk.HTTPMessage{}
	}
	msg.Response = response
	c.messages[response.RequestId] = msg
	c.messageLock.Unlock()
}

// IncRequest count
func (c *Container) IncRequest() {
	atomic.AddInt32(&c.requestCount, 1)
}

// DecRequest count
func (c *Container) DecRequest() {
	atomic.AddInt32(&c.requestCount, -1)
}

// OpenRequestCount number of open requests observed 0 means we have everything
func (c *Container) OpenRequestCount() int32 {
	return atomic.LoadInt32(&c.requestCount)
}

// WaitFor if we have a readyCh for this request, if we don't, make the channel.
// If we do, it is already closed so we can return
func (c *Container) WaitFor(ctx context.Context, requestID string) error {
	var readyCh chan struct{}
	var ok bool

	defer c.remove(requestID)

	c.respReadyLock.Lock()
	if readyCh, ok = c.respReady[requestID]; !ok {
		readyCh = make(chan struct{})
		c.respReady[requestID] = readyCh
	}
	c.respReadyLock.Unlock()

	select {
	case <-readyCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// BodyReady signals WaitFor that the response is done, and we can start reading the body
func (c *Container) BodyReady(requestID string) {
	c.respReadyLock.Lock()
	if _, ok := c.respReady[requestID]; !ok {
		c.respReady[requestID] = make(chan struct{})
	}
	close(c.respReady[requestID])
	c.respReadyLock.Unlock()
}

// remove the request from our ready map
func (c *Container) remove(requestID string) {
	c.respReadyLock.Lock()
	delete(c.respReady, requestID)
	c.respReadyLock.Unlock()
}
