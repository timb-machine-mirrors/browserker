package navi

type Selector struct {
	Query string
	Other string
}

type Find func(s *Selector) (*HTMLElement, error)

type By int8

const (
	ID = iota
	Name
	CSS
	XPath
	Browserk
)

func ByID(s *Selector) (*HTMLElement, error) {
	return nil, nil
}

func ByName(s *Selector) (*HTMLElement, error) {
	return nil, nil
}

func ByCSS(s *Selector) (*HTMLElement, error) {
	return nil, nil
}

func ByXPath(s *Selector) (*HTMLElement, error) {
	return nil, nil
}

func ByBrowserk(s *Selector) (*HTMLElement, error) {
	return nil, nil
}
