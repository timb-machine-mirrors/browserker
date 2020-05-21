package crawler_test

import (
	"testing"

	"gitlab.com/browserker/scanner/crawler"
)

func TestRegexs(t *testing.T) {
	ret := crawler.AddressLine1LabelRe.MatchString("address1")
	if ret == false {
		t.Fatalf("expected match, didn't")
	}
}
