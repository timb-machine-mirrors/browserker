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
	os.RemoveAll("testdata/crawl")
	g := store.NewCrawlGraph("testdata/crawl")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	nav := testMakeNavi([]byte{0, 1, 2})
	nav.OriginID = []byte{}

	if err := g.AddNavigation(nav); err != nil {
		t.Fatalf("error adding: %s\n", err)
	}

	result, err := g.GetNavigation(nav.ID)
	if err != nil {
		t.Fatalf("error reading back navigation: %s\n", err)
	}

	if result.Action == nil {
		t.Fatalf("action was nil")
	}

	if string(nav.ID) != string(result.ID) {
		t.Fatalf("%v != %v\n", nav.ID, result.ID)
	}

	if nav.Action.Type != result.Action.Type {
		t.Fatalf("%v != %v\n", nav.Action.Type, result.Action.Type)
	}
	_ = g.Find(nil, browserker.NavUnvisited, browserker.NavInProcess, 5)
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
		nav.Distance = i - 1

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}

		if err := g.AddNavigation(nav); err != nil {
			t.Fatalf("error adding: %s\n", err)
		}
	}

	limit := 5
	entries := g.Find(nil, browserker.NavUnvisited, browserker.NavUnvisited, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	for i := 0; i < len(entries); i++ {
		t.Logf("%d %d", i, len(entries[i]))
		if len(entries[i]) != i+1 {
			t.Fatalf("entry should have len %d got %d\n", i+1, len(entries[i]))
		}

		last := entries[i][len(entries[i])-1]
		if last.Distance != i {
			t.Fatalf("last entry distance should be %d got %d\n", i, last.Distance)
		}
	}

	entries = g.Find(nil, browserker.NavUnvisited, browserker.NavInProcess, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	for i := 0; i < len(entries); i++ {
		t.Logf("%d %d", i, len(entries[i]))
		if entries[i][0].State != browserker.NavInProcess {
			t.Fatalf("expected in process state was %v\n", entries[i][0].State)
		}
	}
}

func testMakeNavi(id []byte) *browserker.Navigation {
	return &browserker.Navigation{
		ID:               id,
		StateUpdatedTime: time.Now(),
		TriggeredBy:      1,
		State:            browserker.NavUnvisited,
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
