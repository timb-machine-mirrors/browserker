package browserk_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
)

func TestContext(t *testing.T) {
	c := &browserk.Context{}
	count := 5
	hnd := make([]browserk.RequestHandler, count)
	called := 0
	for i := 0; i < count; i++ {
		hnd[i] = func(c *browserk.Context, b browserk.Browser, i *browserk.InterceptedHTTPRequest) {
			called++
			if called == 3 {
				c.ReqAbort()
			}
		}
	}

	c.AddReqHandler(hnd...)
	c.NextReq(nil, nil)
	if called != 3 {
		t.Fatalf("expected abort to kill at 3, got called: %d\n", called)
	}
}
