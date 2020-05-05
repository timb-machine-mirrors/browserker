package navi

type Selector struct {
	Query string
	Other string
}

// revive:disable:var-naming
type Find func(s *Selector) (*Element, error)

type By int8

const (
	ID = iota
	Name
	CSS
	XPath
	Browserk
)

func ByID(s *Selector) (*Element, error) {
	return nil, nil
}

func ByName(s *Selector) (*Element, error) {
	return nil, nil
}

func ByCSS(s *Selector) (*Element, error) {
	return nil, nil
}

func ByXPath(s *Selector) (*Element, error) {
	return nil, nil
}

func ByBrowserk(s *Selector) (*Element, error) {
	return nil, nil
}
