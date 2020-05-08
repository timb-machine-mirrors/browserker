package scanner

import (
	"regexp"
	"strings"

	"gitlab.com/browserker/browserk"
)

type ScopeService struct {
	allowed  *regexp.Regexp
	ignored  *regexp.Regexp
	excluded *regexp.Regexp
}

// TODO differentiate between hosts and urls (for logout etc)
func NewScopeService(allowed, ignored, excluded []string) *ScopeService {
	return &ScopeService{
		allowed:  stringsToRegex(allowed),
		ignored:  stringsToRegex(ignored),
		excluded: stringsToRegex(excluded),
	}
}

func (s *ScopeService) Check(url string) browserk.Scope {
	if s.allowed.MatchString(url) {
		return browserk.In
	} else if s.excluded != nil && s.excluded.MatchString(url) {
		return browserk.Excluded
	}
	return browserk.Out
}

func stringsToRegex(input []string) *regexp.Regexp {
	if input == nil {
		return nil
	}
	merged := strings.Join(input, "|")
	return regexp.MustCompile(merged)
}
