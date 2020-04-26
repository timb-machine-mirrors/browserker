package browserker

import (
	"context"

	"gitlab.com/browserker/browserker/inject"
	"gitlab.com/browserker/browserker/navi"
)

type ActionType int8

const (
	LOAD_URL ActionType = iota
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
)

type Action struct {
	browser   Browser
	Type      ActionType
	Input     []byte
	Result    []byte
	Responses map[int64]*HttpResponse
}

type BrowserOpts struct {
}

type Browser interface {
	ID() int64
	// Load a web page
	Load(ctx context.Context, url string) (err error)
	Find(ctx context.Context, finder navi.Find) (navi.Element, error)
	Instrument(opt *BrowserOpts) error
	InjectBefore(ctx context.Context, inject inject.Injector) error
	InjectAfter(ctx context.Context, inject inject.Injector) ([]byte, error)
	GetResponses() (map[int64]*HttpResponse, error)
	GetRequest() (HttpRequest, error)
	Screenshot(ctx context.Context) ([]byte, error)
	Execute(ctx context.Context, act map[int]*Action) error
	ExecuteSingle(ctx context.Context, act *Action) error
}
