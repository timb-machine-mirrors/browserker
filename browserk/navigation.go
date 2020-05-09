package browserk

import (
	"crypto/md5"
	"time"

	"github.com/rs/zerolog/log"
)

// TriggeredBy stores what caused a navigation attempt
type TriggeredBy int16

const (
	// TrigInitial triggered action (for example start of crawl load url)
	TrigInitial = iota + 1
	// TrigCrawler triggered this navigation
	TrigCrawler
	// TrigPlugin triggered this
	TrigPlugin
	// TrigAutoBrowser something caused the browser to trigger this (redirect etc)
	TrigAutoBrowser
)

// NavState is the state of a navigation
type NavState int8

const (
	// NavInvalid is invalid
	NavInvalid NavState = iota + 1
	// NavUnvisited means it's ready for pick up by crawler
	NavUnvisited
	// NavInProcess crawler is in the process of crawling this
	NavInProcess
	// NavVisited crawler has visited
	NavVisited
	// NavAudited maybe remove, but to set that this navigation has been audited by all plugins
	NavAudited
)

// Navigation for storing the action and results of navigating
type Navigation struct {
	ID               []byte      `graph:"id"`            // cayley does not support []byte keys :|
	OriginID         []byte      `graph:"origin"`        // where this navigation node originated from
	TriggeredBy      TriggeredBy `graph:"trig_by"`       // update to plugin/crawler/manual whatever type
	State            NavState    `graph:"state"`         // state of this navigation
	StateUpdatedTime time.Time   `graph:"state_updated"` // when the state was updated (for timeouts)
	Action           *Action     `graph:"action"`
	Scope            Scope       `graph:"scope"`
	Distance         int         `graph:"dist"`
}

// NewNavigation type
func NewNavigation(triggeredBy TriggeredBy, action *Action) *Navigation {
	n := &Navigation{
		Action:           action,
		TriggeredBy:      triggeredBy,
		State:            NavUnvisited,
		StateUpdatedTime: time.Now(),
	}

	// TODO: add originID as part of new nav id for uniqueness?
	h := md5.New()
	h.Write(n.Action.Input)
	h.Write([]byte{byte(n.Action.Type)})
	n.ID = h.Sum(nil)
	log.Info().Msgf("NEW ID: %#v", n.ID)
	return n
}

// NavigationResult captures result details about a navigation
type NavigationResult struct {
	ID           []byte        `graph:"res_id"`
	NavigationID []byte        `graph:"nav_id"`
	DOM          string        `graph:"dom"`
	StartURL     string        `graph:"start_url"`
	EndURL       string        `graph:"end_url"`
	MessageCount int           `graph:"message_count"`
	Messages     []HTTPMessage `graph:"messages"`
	StorageEvts  []byte
	CookieEvts   []byte
}
