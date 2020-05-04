package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/wirepair/gcd/gcdapi"
	"gitlab.com/browserker/browserker"
	"gitlab.com/browserker/store"
)

func TestCrawlGraph(t *testing.T) {
	g := store.NewCrawlGraph("testdata/crawl")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()
	defer os.RemoveAll("testdata/crawl")

	nav := testMakeNavi([]byte{0, 1, 2})
	nav.OriginID = []byte{0, 0, 0}

	if err := g.AddNavigation(nav); err != nil {
		t.Fatalf("error adding: %s\n", err)
	}

	result, err := g.GetNavigation(nav.ID())
	if err != nil {
		t.Fatalf("error reading back navigation: %s\n", err)
	}

	if result.Action == nil {
		t.Fatalf("action was nil")
	}

	if nav.RequestID != result.RequestID {
		t.Fatalf("%v != %v\n", nav.RequestID, result.RequestID)
	}

	if nav.Action.Type != result.Action.Type {
		t.Fatalf("%v != %v\n", nav.Action.Type, result.Action.Type)
	}
	_ = g.Find(nil, browserker.UNVISITED, browserker.INPROCESS, 5)
}

func TestCrawlAddMultiple(t *testing.T) {
	path := "testdata/multi/crawl"
	os.RemoveAll(path)

	g := store.NewCrawlGraph(path)
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	for i := 1; i < 11; i++ {
		nav := testMakeNavi([]byte{0, byte(i), 2})
		nav.OriginID = []byte{0, byte(i - 1), 2}

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}

		if err := g.AddNavigation(nav); err != nil {
			t.Fatalf("error adding: %s\n", err)
		}
	}

	_ = g.Find(nil, browserker.UNVISITED, browserker.UNVISITED, 5)

}

func testMakeNavi(id []byte) *browserker.Navigation {
	req := testMakeReq()
	return &browserker.Navigation{
		NavigationID: id,
		RequestID:    1,
		DOM:          "blah",
		LoadRequest: &browserker.HTTPRequest{
			NetworkRequest: req,
		},
		Requests: map[int64]*browserker.HTTPRequest{
			0: {
				NetworkRequest: req,
			},
		},
		StateUpdatedTime: time.Now(),
		Responses:        testMakeResp(),
		TriggeredBy:      1,
		State:            browserker.UNVISITED,
		Action: &browserker.Action{
			Type:   browserker.LOAD_URL,
			Input:  nil,
			Result: nil,
		},
	}
}

func testMakeReq() gcdapi.NetworkRequest {
	return gcdapi.NetworkRequest{
		Url:         "http://localhost/",
		UrlFragment: "#spa",
		Method:      "GET",
		Headers: map[string]interface{}{
			"Accept": "something",
		},
		PostData:         "",
		HasPostData:      false,
		MixedContentType: "",
		InitialPriority:  "",
		ReferrerPolicy:   "",
		IsLinkPreload:    false,
	}
}

func testMakeResp() map[int64]*browserker.HTTPResponse {
	return map[int64]*browserker.HTTPResponse{
		0: {
			NetworkResponse: gcdapi.NetworkResponse{
				Url:        "http://localhost/",
				Status:     0,
				StatusText: "",
				Headers: map[string]interface{}{
					"Set-Cookie": "JSESSIONID=123",
				},
				HeadersText: "",
				MimeType:    "",
				RequestHeaders: map[string]interface{}{
					"Accept": "something",
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
				Timing:             nil,
				Protocol:           "",
				SecurityState:      "",
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
		},
	}
}
