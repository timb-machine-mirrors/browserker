package browserk_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
)

func TestUnique(t *testing.T) {
	uniq := browserk.UniqueHost | browserk.UniquePath
	if !uniq.Host() {
		t.Fatalf("expected unique host bit to be set")
	}

	if !uniq.Path() {
		t.Fatalf("expected unique path bit to be set")
	}

	if uniq.Page() {
		t.Fatalf("did not expect unique file bit set")
	}

	if uniq.File() {
		t.Fatalf("did not expect unique file bit set")
	}

	if uniq.Request() {
		t.Fatalf("did not expect unique request bit set")
	}
	if uniq.Response() {
		t.Fatalf("did not expect unique Response bit set")
	}
}
