package crawler

import "gitlab.com/browserker/browserk"

// ElementDiffer is used to capture elements before
// a navigation action, and store their hashes so we can
// remove them as possible candidates after the next action occurs
type ElementDiffer struct {
	elements map[browserk.HTMLElementType]map[string]struct{}
}

// NewElementDiffer container
func NewElementDiffer() *ElementDiffer {
	return &ElementDiffer{
		elements: make(map[browserk.HTMLElementType]map[string]struct{}, 0),
	}
}

// Add a new hash to the element type
func (e *ElementDiffer) Add(element browserk.HTMLElementType, hash []byte) {
	if _, exist := e.elements[element]; !exist {
		e.elements[element] = make(map[string]struct{}, 0)
	}
	e.elements[element][string(hash)] = struct{}{}
}

// Has returns true if element type has a hash equal to hash
func (e *ElementDiffer) Has(element browserk.HTMLElementType, hash []byte) bool {
	if _, exist := e.elements[element]; !exist {
		return false
	}

	_, exist := e.elements[element][string(hash)]
	return exist
}
