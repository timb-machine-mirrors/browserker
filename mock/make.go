package mock

import (
	"fmt"
	"time"

	"github.com/wirepair/gcd/gcdapi"
	"gitlab.com/browserker/browserk"
)

func MakeMockMessages() []*browserk.HTTPMessage {
	m := make([]*browserk.HTTPMessage, 0)
	for i := 0; i < 3; i++ {
		message := &browserk.HTTPMessage{
			RequestTime: time.Now(),
			Request: &browserk.HTTPRequest{
				RequestId:   fmt.Sprintf("%d", i+1),
				LoaderId:    fmt.Sprintf("%d", i+1),
				DocumentURL: fmt.Sprintf("http://example.com/%d", i+1),
				Request: &gcdapi.NetworkRequest{
					Url:         fmt.Sprintf("http://example.com/%d", i+1),
					UrlFragment: "",
					Method:      "GET",
					Headers: map[string]interface{}{
						"": nil,
					},
					PostData:         "",
					HasPostData:      false,
					MixedContentType: "",
					InitialPriority:  "",
					ReferrerPolicy:   "",
					IsLinkPreload:    false,
				},
				Timestamp:        0.0,
				WallTime:         0.0,
				Initiator:        nil,
				RedirectResponse: nil,
			},
			RequestMod:   nil,
			ResponseTime: time.Now().Add(time.Second * 3),
			Response: &browserk.HTTPResponse{
				RequestId: fmt.Sprintf("%d", i+1),
				LoaderId:  fmt.Sprintf("%d", i+1),
				Timestamp: 0.0,
				Type:      "Document",
				Response: &gcdapi.NetworkResponse{
					Url:        fmt.Sprintf("http://example.com/%d", i+1),
					Status:     200,
					StatusText: "OK",
					Headers: map[string]interface{}{
						"": nil,
					},
					HeadersText: "",
					MimeType:    "text/html",
					RequestHeaders: map[string]interface{}{
						"": nil,
					},
					RequestHeadersText: "",
					ConnectionReused:   false,
					ConnectionId:       0.0,
					RemoteIPAddress:    "",
					RemotePort:         0,
					FromDiskCache:      false,
					FromServiceWorker:  false,
					FromPrefetchCache:  false,
					EncodedDataLength:  0.0,
					Timing: &gcdapi.NetworkResourceTiming{
						RequestTime:       0.0,
						ProxyStart:        0.0,
						ProxyEnd:          0.0,
						DnsStart:          0.0,
						DnsEnd:            0.0,
						ConnectStart:      0.0,
						ConnectEnd:        0.0,
						SslStart:          0.0,
						SslEnd:            0.0,
						WorkerStart:       0.0,
						WorkerReady:       0.0,
						SendStart:         0.0,
						SendEnd:           0.0,
						PushStart:         0.0,
						PushEnd:           0.0,
						ReceiveHeadersEnd: 0.0,
					},
					Protocol:      "",
					SecurityState: "",
					SecurityDetails: &gcdapi.NetworkSecurityDetails{
						Protocol:                          "",
						KeyExchange:                       "",
						KeyExchangeGroup:                  "",
						Cipher:                            "",
						Mac:                               "",
						CertificateId:                     0,
						SubjectName:                       "",
						SanList:                           nil,
						Issuer:                            "",
						ValidFrom:                         0.0,
						ValidTo:                           0.0,
						SignedCertificateTimestampList:    nil,
						CertificateTransparencyCompliance: "",
					},
				},
				FrameId:  "",
				Body:     nil,
				BodyHash: nil,
			},
		}
		m = append(m, message)
	}
	return m
}

func MakeMockCookies() []*browserk.Cookie {
	c := make([]*browserk.Cookie, 0)
	for i := 0; i < 3; i++ {
		c = append(c, &browserk.Cookie{
			Name:         fmt.Sprintf("name%d", i+1),
			Value:        fmt.Sprintf("value%d", i+1),
			Domain:       "",
			Path:         "",
			Expires:      0.0,
			Size:         0,
			HTTPOnly:     true,
			Secure:       true,
			Session:      true,
			SameSite:     "",
			Priority:     "",
			ObservedTime: time.Time{},
		})
	}
	return c
}

func MakeMockConsole() []*browserk.ConsoleEvent {
	c := make([]*browserk.ConsoleEvent, 0)
	for i := 0; i < 3; i++ {
		c = append(c, &browserk.ConsoleEvent{
			Source:   "",
			Level:    "",
			Text:     fmt.Sprintf("name%d", i+1),
			URL:      "",
			Line:     0,
			Column:   0,
			Observed: time.Now().Add(time.Second * time.Duration(i)),
		})
	}
	return c
}

func MakeMockStorage() []*browserk.StorageEvent {
	s := make([]*browserk.StorageEvent, 0)
	for i := 0; i < 3; i++ {
		s = append(s, &browserk.StorageEvent{
			Type:           browserk.StorageAddedEvt,
			IsLocalStorage: false,
			SecurityOrigin: "",
			Key:            fmt.Sprintf("key%d", i+1),
			NewValue:       fmt.Sprintf("value%d", i+1),
			OldValue:       "",
			Observed:       time.Now().Add(time.Second * time.Duration(i)),
		})
	}
	return s
}

func MakeMockResult(id []byte) *browserk.NavigationResult {
	r := &browserk.NavigationResult{
		NavigationID:  id,
		DOM:           "<html>nav result</html>",
		StartURL:      "http://example.com/start" + fmt.Sprintf("%x", id),
		EndURL:        "http://example.com/end",
		MessageCount:  1,
		Messages:      MakeMockMessages(),
		Cookies:       MakeMockCookies(),
		ConsoleEvents: MakeMockConsole(),
		StorageEvents: MakeMockStorage(),
		CausedLoad:    false,
		WasError:      false,
		Errors:        nil,
	}
	r.Hash()
	return r
}

func MakeMockNavi(id []byte) *browserk.Navigation {
	return &browserk.Navigation{
		ID:               id,
		StateUpdatedTime: time.Now(),
		TriggeredBy:      1,
		State:            browserk.NavUnvisited,
		Action: &browserk.Action{
			Type:   browserk.ActLoadURL,
			Input:  nil,
			Result: nil,
		},
	}
}
