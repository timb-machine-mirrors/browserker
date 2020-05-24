package crawler_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
)

var leaser = browser.NewLocalLeaser()

func testServer(path string, fn gin.HandlerFunc) (string, *http.Server) {
	router := gin.Default()
	router.Static("/forms", "testdata/forms")
	if fn != nil {
		router.Any(path, fn)
	}
	testListener, _ := net.Listen("tcp", ":0")
	_, testServerPort, _ := net.SplitHostPort(testListener.Addr().String())
	srv := &http.Server{
		Addr:    testListener.Addr().String(),
		Handler: router,
	}
	//testServerAddr := fmt.Sprintf("http://localhost:%s/", testServerPort)
	go func() {
		if err := srv.Serve(testListener); err != http.ErrServerClosed {
			log.Fatalf("Serve(): %s", err)
		}
	}()

	return testServerPort, srv
}

func TestCrawler(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()

	ctx := context.Background()
	bCtx := mock.Context(ctx)
	bCtx.FormHandler = crawler.NewCrawlerFormHandler(&browserk.DefaultFormValues)

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	called := false
	formHandler := func(c *gin.Context) {
		called = true
		resp := "<html><body>You made it!</body></html>"
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(resp))
	}

	p, srv := testServer("/result/addAddress", formHandler)
	defer srv.Shutdown(ctx)

	crawl := crawler.New(&browserk.Config{})

	target := fmt.Sprintf("http://localhost:%s/", p)
	navURL := target + "forms/"

	targetURL, _ := url.Parse(target)
	bCtx.Scope = scanner.NewScopeService(targetURL)

	act := browserk.NewLoadURLAction(navURL)
	nav := browserk.NewNavigation(browserk.TrigCrawler, act)
	_, newNavs, err := crawl.Process(bCtx, b, nav, true)
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}

	if len(newNavs) != 1 {
		t.Fatal("did not find form nav action")
	}
	_, _, err = crawl.Process(bCtx, b, newNavs[0], true)
	if err != nil {
		t.Fatalf("failed to submit form %s\n", err)
	}

	if !called {
		t.Fatalf("form was not submitted")
	}
}
