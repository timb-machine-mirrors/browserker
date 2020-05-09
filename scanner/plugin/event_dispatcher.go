package plugin

import "gitlab.com/browserker/browserk"

type EventDispatcher struct {
	eventCh chan struct{}
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{}
}

// Listener dispatches event to all plugins that registered for the event type
func (e *EventDispatcher) Listener(c *browserk.Context) {
}
