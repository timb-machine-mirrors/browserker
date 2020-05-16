package scanner_test

import (
	"net/url"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
)

func TestScope(t *testing.T) {
	target, _ := url.Parse("http://example.com")

	allowed := []string{"example.com"}
	ignored := []string{"bad.com"}
	s := scanner.NewScopeService(target)
	s.AddScope(allowed, browserk.InScope)
	s.AddScope(ignored, browserk.OutOfScope)
	s.AddExcludedURIs([]string{"/log-out", "/signout"})

	var inputs = []struct {
		in       string
		expected browserk.Scope
	}{
		{
			"http://example.com",
			browserk.InScope,
		}, {
			"http://bad.com",
			browserk.OutOfScope,
		},
		{
			"http://example.com/bad.com",
			browserk.InScope,
		},
		{
			"https://bad.com/example.com",
			browserk.OutOfScope,
		},
		{
			"http://example.com/log-out",
			browserk.ExcludedFromScope,
		},
		{
			"http://example.com/signout",
			browserk.ExcludedFromScope,
		},
		{
			"http://example.com/different/signout",
			browserk.InScope,
		},
		{
			"http://bad.com/signout",
			browserk.OutOfScope,
		},
	}
	for _, in := range inputs {
		ret := s.Check(in.in)
		if ret != in.expected {
			t.Fatalf("%v did not match %v for %s\n", ret, in.expected, in.in)
		}
	}
}
