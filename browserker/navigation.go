package browserker

type NavNode interface {
	ID()
}

type NavEdge interface {
	ID()
	Next()
}

type NavState int8

const (
	UNVISITED = iota // UNVISITED means it's ready for pick up by crawler
	INPROCESS        // crawler is in the process of crawling this
	VISITED
	AUDITED
)

type Navigation struct {
	NavigationID int64
	RequestID    int64
	DOM          string
	LoadRequest  *HTTPRequest
	Requests     map[string]*HTTPRequest
	Responses    map[string]*HTTPResponse
	TriggeredBy  int      // update to plugin/crawler/manual whatever type
	State        NavState // state of this navigation
	Action       Action
	ActionType
}
