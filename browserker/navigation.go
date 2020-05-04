package browserker

import (
	"crypto/md5"
	"time"
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
	INVALID_STATE NavState = iota + 1
	// UNVISITED means it's ready for pick up by crawler
	UNVISITED
	// INPROCESS crawler is in the process of crawling this
	INPROCESS
	// VISITED crawler has visited
	VISITED
	// AUDITED maybe remove, but to set that this navigation has been audited by all plugins
	AUDITED
)

// Navigation for storing the action and results of navigating
type Navigation struct {
	NavigationID     []byte      `graph:"id"`     // cayley does not support []byte keys :|
	OriginID         []byte      `graph:"origin"` // where this navigation node originated from
	RequestID        int64       `graph:"request_id"`
	TriggeredBy      TriggeredBy `graph:"trig_by"`       // update to plugin/crawler/manual whatever type
	State            NavState    `graph:"state"`         // state of this navigation
	StateUpdatedTime time.Time   `graph:"state_updated"` // when the state was updated (for timeouts)
	Action           *Action     `graph:"action"`

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
	n.NavigationID = md5.New().Sum(append(n.Action.Input, byte(n.Action.Type)))
	return n
}

// ID returns the hash of action input and type
func (n *Navigation) ID() []byte {
	return n.NavigationID
}
