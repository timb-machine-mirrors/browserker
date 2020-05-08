package browserk_test

import (
	"fmt"
	"testing"

	"gitlab.com/browserker/browserk"
)

func TestContext(t *testing.T) {
	c := &browserk.Context{}
	count := 5
	hnd := make([]browserk.RequestHandler, count)
	called := 0
	for i := 0; i < count; i++ {
		hnd[i] = func(c *browserk.Context) {
			fmt.Printf("%d\n", called)
			called++
			if called == 3 {
				c.ReqAbort()
			}
		}
	}

	c.AddReqHandler(hnd...)
	c.NextReq()
	if called != 3 {
		t.Fatalf("expected abort to kill at 3, got called: %d\n", called)
	}
}
