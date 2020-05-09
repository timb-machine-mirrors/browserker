package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
	"gitlab.com/browserker/browserk"
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

func (t *Tab) subscribeNetworkEvents(ctx *browserk.Context) {
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
			log.Info().Str("request_id", message.Params.RequestId).Msg("is Document request")
			t.container.SetLoadRequest(req)
		}
		t.container.AddRequest(req)
		log.Info().Str("request_id", message.Params.RequestId).Msg("added request")
	})

	t.t.Subscribe("Network.responseReceived", func(target *gcd.ChromeTarget, payload []byte) {
		defer t.container.DecRequest()

		message := &gcdapi.NetworkResponseReceivedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			return
		}
		p := message.Params
		log.Info().Str("request_id", message.Params.RequestId).Msg("waiting")

		timeoutCtx, cancel := context.WithTimeout(ctx.Ctx, time.Second*60)
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
		t.container.AddResponse(GCDResponseToBrowserk(message, body))
		log.Info().Msg("added")

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

func (t *Tab) subscribeInterception(ctx *browserk.Context) {
	t.t.Subscribe("Fetch.requestPaused", func(target *gcd.ChromeTarget, payload []byte) {
		log.Info().Msg("Fetch.requestPaused event")

		message := &gcdapi.FetchRequestPausedEvent{}
		if err := json.Unmarshal(payload, message); err != nil {
			log.Fatal().Err(err).Msg("critical error Fetch.requestPaused event was unable to decode")
		}

		p := message.Params
		// we are in a response paused event
		if p.ResponseHeaders != nil {
			bodyStr, encoded, err := t.t.Fetch.GetResponseBody(p.RequestId)
			if err != nil {
				log.Warn().Err(err).Msg("unable to get body")
				respParams := &gcdapi.FetchFulfillRequestParams{
					RequestId: p.RequestId,
				}
				t.t.Fetch.FulfillRequestWithParams(respParams)
				return
			}

			modified := GCDFetchResponseToIntercepted(message, bodyStr, encoded)

			ctx.NextResp(t, modified)
			respParams := &gcdapi.FetchFulfillRequestParams{
				RequestId: p.RequestId,
			}

			if modified.Modified.ResponseCode != 0 {
				respParams.ResponseCode = modified.Modified.ResponseCode
			} else {
				respParams.ResponseCode = p.ResponseStatusCode
			}
			if modified.Modified.Body != nil {
				respParams.Body = base64.StdEncoding.EncodeToString(modified.Modified.Body)
			}

			if modified.Modified.ResponseHeaders != nil {
				respParams.ResponseHeaders = modified.Modified.ResponseHeaders
			}
			if modified.Modified.ResponsePhrase != "" {
				respParams.ResponsePhrase = modified.Modified.ResponsePhrase
			}
			t.t.Fetch.FulfillRequestWithParams(respParams)
		} else {
			// we are in a request paused event
			modified := GCDFetchRequestToIntercepted(message, t.container)
			log.Info().Msg("Calling request Hooks")
			ctx.NextReq(t, modified)

			reqParams := &gcdapi.FetchContinueRequestParams{
				RequestId: modified.RequestId,
			}

			if modified.Modified.Method != "" {
				reqParams.Method = modified.Modified.Method
			}
			if modified.Modified.Url != "" {
				reqParams.Url = modified.Modified.Url
			}
			if modified.Modified.Headers != nil {
				reqParams.Headers = modified.Modified.Headers
			}
			if modified.Modified.PostData != "" {
				reqParams.PostData = modified.Modified.PostData
			}
			t.t.Fetch.ContinueRequestWithParams(reqParams)
		}

	})
}
