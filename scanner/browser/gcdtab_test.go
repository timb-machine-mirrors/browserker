package browser_test

import (
	"log"
	"net"
	"net/http"
	"testing"

	"gitlab.com/browserker/scanner/browser"
	"golang.org/x/net/context"
)

var leaser = browser.NewLocalLeaser()

func testServer() (string, *http.Server) {
	srv := &http.Server{Handler: http.FileServer(http.Dir("testdata/"))}
	testListener, _ := net.Listen("tcp", ":0")
	_, testServerPort, _ := net.SplitHostPort(testListener.Addr().String())
	//testServerAddr := fmt.Sprintf("http://localhost:%s/", testServerPort)
	go func() {
		if err := srv.Serve(testListener); err != http.ErrServerClosed {
			log.Fatalf("Serve(): %s", err)
		}
	}()

	return testServerPort, srv
}

func TestStartBrowsers(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()

	ctx := context.Background()
	b, err := pool.Take(ctx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}
	b.Navigate(ctx, "http://example.com")
	t.Logf(b.SerializeDOM())
}
