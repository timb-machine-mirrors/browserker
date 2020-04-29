package browserker

import (
	"context"

	"gitlab.com/browserker/browserker/inject"
	"gitlab.com/browserker/browserker/navi"
)

// ActionType defines the action type for a browser action
type ActionType int8

// revive:disable:var-naming
const (
	LOAD_URL ActionType = iota + 1
	EXECUTE_JS
	LEFTCLICK
	LEFTCLICK_DOWN
	LEFTCLICK_UP
	RIGHTCLICK
	RIGHTCLICK_DOWN
	RIGHTCLICK_UP
	MIDDLECLICK
	MIDDLECLICK_DOWN
	MIDDLECLICK_UP
	SCROLL
	SENDKEYS
	KEYUP
	KEYDOWN
	HOVER
	FOCUS
	WAIT

	// ActionTypes that occured automatically
	REDIRECT
	SUB_REQUEST
)

// Action runs a browser action
type Action struct {
	browser Browser
	Type    ActionType `quad:"type"`
	Input   []byte     `quad:"input"`
	Result  []byte     `quad:"input"`
}

// BrowserOpts todo: define
type BrowserOpts struct {
}

// Browser interface
type Browser interface {
	ID() int64
	// Load a web page
	Load(ctx context.Context, url string) (err error)
	Find(ctx context.Context, finder navi.Find) (navi.Element, error)
	Instrument(opt *BrowserOpts) error
	InjectBefore(ctx context.Context, inject inject.Injector) error
	InjectAfter(ctx context.Context, inject inject.Injector) ([]byte, error)
	GetResponses() (map[int64]*HTTPResponse, error)
	GetRequest() (HTTPRequest, error)
	Screenshot(ctx context.Context) ([]byte, error)
	Execute(ctx context.Context, act map[int]*Action) error
	ExecuteSingle(ctx context.Context, act *Action) error
}
