package browser_test

import (
	"testing"

	"github.com/wirepair/gcd/gcdapi"
	"gitlab.com/browserker/scanner/browser"
)

func TestNodeRemove(t *testing.T) {
	d := &gcdapi.DOMNode{
		Attributes: []string{"href", "blah", "zup", "zop"},
	}
	ret := browser.NodeRemoveAttribute(d, "zup")
	if len(ret) != 2 {
		t.Fatalf("expected 2 left")
	}
	if ret[0] != "href" && ret[1] != "blah" {
		t.Fatalf("excepcted href=blah left")
	}
}
