package browserk

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
	ActFillForm

	// ActionTypes that occured automatically
	ActRedirect
	ActSubRequest
)

// Action runs a browser action, may or may not create a result
type Action struct {
	browser Browser
	Type    ActionType       `graph:"type"`
	Input   []byte           `graph:"input"`
	Element *HTMLElement     `graph:"element"`
	Form    *HTMLFormElement `graph:"form"`
	Result  []byte           `graph:"result"`
}
