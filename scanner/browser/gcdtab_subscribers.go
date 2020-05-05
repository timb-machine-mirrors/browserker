package browser

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

func (t *Tab) subscribeTargetCrashed() {
	t.t.Subscribe("Inspector.targetCrashed", func(target *gcd.ChromeTarget, payload []byte) {
		select {
		case t.crashedCh <- "crashed":
		case <-t.exitCh:
		}
	})
}

func (t *Tab) subscribeTargetDetached() {
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
}

// our default loadFiredEvent handler, returns a response to resp channel to navigate once complete.
func (t *Tab) subscribeLoadEvent() {
	t.t.Subscribe("Page.loadEventFired", func(target *gcd.ChromeTarget, payload []byte) {
		if t.IsNavigating() {
			select {
			case t.navigationCh <- 0:
			case <-t.exitCh:
			}

		}
	})
}

func (t *Tab) subscribeFrameLoadingEvent() {
	t.t.Subscribe("Page.frameStartedLoading", func(target *gcd.ChromeTarget, payload []byte) {
		log.Debug().Msgf("frameStartedLoading: %s\n", string(payload))
		if t.IsNavigating() {
			return
		}
		header := &gcdapi.PageFrameStartedLoadingEvent{}
		err := json.Unmarshal(payload, header)
		// has the top frame id begun navigating?
		if err == nil && header.Params.FrameId == t.GetTopFrameID() {
			t.setIsTransitioning(true)
		}
	})
}

func (t *Tab) subscribeFrameFinishedEvent() {
	t.t.Subscribe("Page.frameStoppedLoading", func(target *gcd.ChromeTarget, payload []byte) {
		log.Debug().Msgf("frameStoppedLoading: %s\n", string(payload))
		if t.IsNavigating() {
			return
		}
		header := &gcdapi.PageFrameStoppedLoadingEvent{}
		err := json.Unmarshal(payload, header)
		// has the top frame id begun navigating?
		if err == nil && header.Params.FrameId == t.GetTopFrameID() {
			t.setIsTransitioning(false)
		}
	})
}

func (t *Tab) subscribeSetChildNodes() {
	// new nodes
	t.t.Subscribe("DOM.setChildNodes", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMSetChildNodesEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: SetChildNodesEvent, Nodes: event.Nodes, ParentNodeId: event.ParentId})

		}
	})
}

func (t *Tab) subscribeAttributeModified() {
	t.t.Subscribe("DOM.attributeModified", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMAttributeModifiedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: AttributeModifiedEvent, Name: event.Name, Value: event.Value, NodeId: event.NodeId})
		}
	})
}

func (t *Tab) subscribeAttributeRemoved() {
	t.t.Subscribe("DOM.attributeRemoved", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMAttributeRemovedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: AttributeRemovedEvent, NodeId: event.NodeId, Name: event.Name})
		}
	})
}
func (t *Tab) subscribeCharacterDataModified() {
	t.t.Subscribe("DOM.characterDataModified", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMCharacterDataModifiedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: CharacterDataModifiedEvent, NodeId: event.NodeId, CharacterData: event.CharacterData})
		}
	})
}
func (t *Tab) subscribeChildNodeCountUpdated() {
	t.t.Subscribe("DOM.childNodeCountUpdated", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMChildNodeCountUpdatedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: ChildNodeCountUpdatedEvent, NodeId: event.NodeId, ChildNodeCount: event.ChildNodeCount})
		}
	})
}
func (t *Tab) subscribeChildNodeInserted() {
	t.t.Subscribe("DOM.childNodeInserted", func(target *gcd.ChromeTarget, payload []byte) {
		//log.Printf("childNodeInserted: %s\n", string(payload))
		header := &gcdapi.DOMChildNodeInsertedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: ChildNodeInsertedEvent, Node: event.Node, ParentNodeId: event.ParentNodeId, PreviousNodeId: event.PreviousNodeId})
		}
	})
}
func (t *Tab) subscribeChildNodeRemoved() {
	t.t.Subscribe("DOM.childNodeRemoved", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMChildNodeRemovedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: ChildNodeRemovedEvent, ParentNodeId: event.ParentNodeId, NodeId: event.NodeId})
		}
	})
}

func (t *Tab) dispatchNodeChange(evt *NodeChangeEvent) {
	select {
	case t.nodeChange <- evt:
	case <-t.exitCh:
		return
	}
}

/*
func (t *Tab) subscribeInlineStyleInvalidated() {
	t.t.Subscribe("DOM.inlineStyleInvalidatedEvent", func(target *gcd.ChromeTarget, payload []byte) {
		event := &gcdapi.DOMInlineStyleInvalidatedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event = header.Params
			t.nodeChange <- &NodeChangeEvent{EventType: InlineStyleInvalidatedEvent, NodeIds: event.NodeIds}
		}
	})
}
*/
func (t *Tab) subscribeDocumentUpdated() {
	// node ids are no longer valid
	t.t.Subscribe("DOM.documentUpdated", func(target *gcd.ChromeTarget, payload []byte) {
		select {
		case t.nodeChange <- &NodeChangeEvent{EventType: DocumentUpdatedEvent}:
		case <-t.exitCh:
		}
	})
}
