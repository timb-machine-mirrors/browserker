package browserk

import (
	"context"
)

// ActionType defines the action type for a browser action
type ActionType int8

// revive:disable:var-naming
const (
	ActLoadURL ActionType = iota + 1
	ActExecuteJS
	ActLeftClick
	ActLeftClickDown
	ActLeftClickUp
	ActRightClick
	ActRightClickDown
	ActRightClickUp
	ActMiddleClick
	ActMiddleClickDown
	ActMiddleClickUp
	ActScroll
	ActSendKeys
	ActKeyUp
	ActKeyDown
	ActHover
	ActFocus
	ActWait

	// ActionTypes that occured automatically
	ActRedirect
	ActSubRequest
)

// Action runs a browser action
type Action struct {
	browser Browser
	Type    ActionType `graph:"type"`
	Input   []byte     `graph:"input"`
	Result  []byte     `graph:"result"`
}

// BrowserPool handles taking/returning browsers
type BrowserPool interface {
	Take(ctx *Context) (Browser, error)
	Return(ctx context.Context, browserPort string)
	Leased() int
	Shutdown() error
}

// BrowserOpts todo: define
type BrowserOpts struct {
}

// Browser interface
type Browser interface {
	ID() int64
	// Navigate to a web page
	Navigate(ctx context.Context, url string) (err error)
	Find(ctx context.Context, finder Find) (*HTMLElement, error)
	Instrument(opt *BrowserOpts) error
	GetMessages() ([]*HTTPMessage, error)
	Screenshot(ctx context.Context) (string, error)
	Execute(ctx context.Context, act map[int]*Action) error
	ExecuteSingle(ctx context.Context, act *Action) error
}
