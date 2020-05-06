package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/davecgh/go-spew/spew"
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
		if err == nil && header.Params.FrameId == t.getTopFrameID() {
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
		if err == nil && header.Params.FrameId == t.getTopFrameID() {
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
			t.dispatchNodeChange(&NodeChangeEvent{EventType: SetChildNodesEvent, Nodes: event.Nodes, ParentNodeID: event.ParentId})

		}
	})
}

func (t *Tab) subscribeAttributeModified() {
	t.t.Subscribe("DOM.attributeModified", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMAttributeModifiedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: AttributeModifiedEvent, Name: event.Name, Value: event.Value, NodeID: event.NodeId})
		}
	})
}

func (t *Tab) subscribeAttributeRemoved() {
	t.t.Subscribe("DOM.attributeRemoved", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMAttributeRemovedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: AttributeRemovedEvent, NodeID: event.NodeId, Name: event.Name})
		}
	})
}
func (t *Tab) subscribeCharacterDataModified() {
	t.t.Subscribe("DOM.characterDataModified", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMCharacterDataModifiedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: CharacterDataModifiedEvent, NodeID: event.NodeId, CharacterData: event.CharacterData})
		}
	})
}
func (t *Tab) subscribeChildNodeCountUpdated() {
	t.t.Subscribe("DOM.childNodeCountUpdated", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMChildNodeCountUpdatedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: ChildNodeCountUpdatedEvent, NodeID: event.NodeId, ChildNodeCount: event.ChildNodeCount})
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
			t.dispatchNodeChange(&NodeChangeEvent{EventType: ChildNodeInsertedEvent, Node: event.Node, ParentNodeID: event.ParentNodeId, PreviousNodeID: event.PreviousNodeId})
		}
	})
}
func (t *Tab) subscribeChildNodeRemoved() {
	t.t.Subscribe("DOM.childNodeRemoved", func(target *gcd.ChromeTarget, payload []byte) {
		header := &gcdapi.DOMChildNodeRemovedEvent{}
		err := json.Unmarshal(payload, header)
		if err == nil {
			event := header.Params
			t.dispatchNodeChange(&NodeChangeEvent{EventType: ChildNodeRemovedEvent, ParentNodeID: event.ParentNodeId, NodeID: event.NodeId})
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

func (t *Tab) subscribeNetworkEvents(ctx context.Context) {
	t.t.Subscribe("network.loadingFailed", func(target *gcd.ChromeTarget, payload []byte) {
		log.Info().Msgf("failed: %s\n", string(payload))
		t.container.DecRequest()
	})

	t.t.Subscribe("Network.requestWillBeSent", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.NetworkRequestWillBeSentEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		log.Info().Str("request_id", message.Params.RequestId).Msg("adding request")
		req := GCDRequestToBrowserk(message)
		t.container.IncRequest()

		if message.Params.Type == "Document" {
			t.container.SetLoadRequest(req)
		}
		t.container.AddRequest(req)
	})

	t.t.Subscribe("Network.responseReceived", func(target *gcd.ChromeTarget, payload []byte) {
		defer t.container.DecRequest()

		message := &gcdapi.NetworkResponseReceivedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		p := message.Params
		log.Info().Str("request_id", message.Params.RequestId).Msg("waiting")

		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*60)
		defer cancel()

		// log.Info().Str("request_id", p.RequestId).Msg("waiting")
		if err := t.container.WaitFor(timeoutCtx, p.RequestId); err != nil {
			return
		}
		bodyStr, encoded, err := t.t.Network.GetResponseBody(message.Params.RequestId)
		if err != nil {
			log.Warn().Str("url", message.Params.Response.Url).Err(err).Msg("failed to get body")
		}

		body := []byte(bodyStr)
		if encoded {
			body, _ = base64.StdEncoding.DecodeString(bodyStr)
		}
		log.Info().Msg("adding response w/body to container")
		spew.Dump(body)

		t.container.AddResponse(GCDResponseToBrowserk(message, body))
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

func (t *Tab) subscribeStorageEvents(storageFn StorageFunc) {
	t.t.Subscribe("Storage.domStorageItemsCleared", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.DOMStorageDomStorageItemsClearedEvent{}
		if err := json.Unmarshal(payload, message); err == nil {
			p := message.Params
			storageEvent := &StorageEvent{IsLocalStorage: p.StorageId.IsLocalStorage, SecurityOrigin: p.StorageId.SecurityOrigin}
			storageFn(t, "cleared", storageEvent)
		}
	})
	t.t.Subscribe("Storage.domStorageItemRemoved", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.DOMStorageDomStorageItemRemovedEvent{}
		if err := json.Unmarshal(payload, message); err == nil {
			p := message.Params
			storageEvent := &StorageEvent{IsLocalStorage: p.StorageId.IsLocalStorage, SecurityOrigin: p.StorageId.SecurityOrigin, Key: p.Key}
			storageFn(t, "removed", storageEvent)
		}
	})
	t.t.Subscribe("Storage.domStorageItemAdded", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.DOMStorageDomStorageItemAddedEvent{}
		if err := json.Unmarshal(payload, message); err == nil {
			p := message.Params
			storageEvent := &StorageEvent{IsLocalStorage: p.StorageId.IsLocalStorage, SecurityOrigin: p.StorageId.SecurityOrigin, Key: p.Key, NewValue: p.NewValue}
			storageFn(t, "added", storageEvent)
		}
	})
	t.t.Subscribe("Storage.domStorageItemUpdated", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.DOMStorageDomStorageItemUpdatedEvent{}
		if err := json.Unmarshal(payload, message); err == nil {
			p := message.Params
			storageEvent := &StorageEvent{IsLocalStorage: p.StorageId.IsLocalStorage, SecurityOrigin: p.StorageId.SecurityOrigin, Key: p.Key, NewValue: p.NewValue, OldValue: p.OldValue}
			storageFn(t, "updated", storageEvent)
		}
	})
}

func (t *Tab) subscribeInterception(ctx context.Context) {
	t.t.Subscribe("Fetch.requestPaused", func(target *gcd.ChromeTarget, payload []byte) {
		message := &gcdapi.FetchRequestPausedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			log.Fatal().Err(err).Msg("critical error Fetch.requestPaused event was unable to decode")
		}

		p := message.Params
		if p.ResponseHeaders != nil {
			bodyStr, encoded, err := t.t.Fetch.GetResponseBody(p.RequestId)
			if err != nil {
				log.Info().Err(err).Msg("error fetch GetResponseBody")
			} else {
				body := []byte(bodyStr)
				if encoded {
					body, _ = base64.StdEncoding.DecodeString(bodyStr)
				}
				log.Info().Msg("dumping response from GetResponseBody InFetch")
				spew.Dump(body)
				t.t.Fetch.FulfillRequestWithParams(&gcdapi.FetchFulfillRequestParams{
					RequestId:    p.RequestId,
					ResponseCode: 200,
					Body:         base64.StdEncoding.EncodeToString(body),
				})
			}

			log.Info().Msg("ContinueRequestWithParams REQUEST DATA")
			//time.Sleep(5 * time.Second)
			t.t.Fetch.ContinueRequestWithParams(&gcdapi.FetchContinueRequestParams{
				RequestId: p.RequestId,
				Url:       p.Request.Url,
				Method:    p.Request.Method,
			})

		}
	})
}
