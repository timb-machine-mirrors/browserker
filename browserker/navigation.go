package browserker

import (
	"crypto/md5"
	"time"

	"github.com/gobuffalo/packr/v2/file/resolver/encoding/hex"
)

type NavNode interface {
	ID() string
}

type NavEdge interface {
	ID()
	Next()
}

// TriggeredBy stores what caused a navigation attempt
type TriggeredBy int16

const (
	// INITIAL triggered action (for example start of crawl load url)
	INITIAL = iota + 1
	// CRAWLER triggered this navigation
	CRAWLER
	// PLUGIN triggered this
	PLUGIN
	// AUTO_BROWSER something caused the browser to trigger this (redirect etc)
	AUTO_BROWSER
)

type NavState int8

const (
	// UNVISITED means it's ready for pick up by crawler
	UNVISITED = iota + 1
	// INPROCESS crawler is in the process of crawling this
	INPROCESS
	// VISITED crawler has visited
	VISITED
	// AUDITED maybe remove, but to set that this navigation has been audited by all plugins
	AUDITED
)

// Navigation for storing the action and results of navigating
type Navigation struct {
	NavigationID     string      `quad:"@id"`    // cayley does not support []byte keys :|
	OriginID         string      `quad:"origin"` // where this navigation node originated from
	RequestID        int64       `quad:"nav_id"`
	TriggeredBy      TriggeredBy `quad:"trig_by"`       // update to plugin/crawler/manual whatever type
	State            NavState    `quad:"state"`         // state of this navigation
	StateUpdatedTime time.Time   `quad:"state_updated"` // when the state was updated (for timeouts)
	Action           *Action     `quad:"action"`

	// todo maybe make a navigation result with these details?
	DOM         string
	LoadRequest *HTTPRequest
	Requests    map[int64]*HTTPRequest
	Responses   map[int64]*HTTPResponse
}

// NewNavigation type
func NewNavigation(triggeredBy TriggeredBy, action *Action) *Navigation {
	n := &Navigation{
		Action:      action,
		TriggeredBy: triggeredBy,
	}
	n.NavigationID = hex.EncodeToString(md5.New().Sum(append(n.Action.Input, byte(n.Action.Type))))
	return n
}

// ID returns the hash of action input and type
func (n *Navigation) ID() string {
	return string(n.NavigationID)
}
