package injections

import "gitlab.com/browserker/scanner/injections/injast"

type Cursor struct {
	// contains filtered or unexported fields
}

type ApplyFunc func(*Cursor) bool

// TODO: Implement https://godoc.org/golang.org/x/tools/go/ast/astutil#Apply style rewriting
// taking care of encoding of nested types (eg. xml inside of json)
type Injector interface {
	Apply(root injast.Node, pre, post ApplyFunc) (result injast.Node)
}
