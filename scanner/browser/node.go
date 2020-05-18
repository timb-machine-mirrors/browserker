package browser

import (
	"strings"

	"github.com/wirepair/gcd/gcdapi"
)

func NodeHasAttribute(node *gcdapi.DOMNode, attr string) bool {
	attr = strings.ToLower(attr)
	if node.Attributes == nil {
		node.Attributes = make([]string, 0)
	}
	for i := 0; i < len(node.Attributes); i += 2 {
		if strings.ToLower(node.Attributes[i]) == attr {
			return true
		}
	}
	return false
}

func NodeGetAttribute(node *gcdapi.DOMNode, attr string) string {
	attr = strings.ToLower(attr)
	if node.Attributes == nil {
		node.Attributes = make([]string, 0)
	}
	for i := 0; i < len(node.Attributes); i += 2 {
		if strings.ToLower(node.Attributes[i]) == attr {
			return node.Attributes[i+1]
		}
	}
	return ""
}

func NodeUpdateAttribute(node *gcdapi.DOMNode, attr, value string) {
	attr = strings.ToLower(attr)
	if node.Attributes == nil {
		node.Attributes = make([]string, 0)
	}
	for i := 0; i < len(node.Attributes); i += 2 {
		if strings.ToLower(node.Attributes[i]) == attr {
			node.Attributes[i+1] = value
			return
		}
	}
	// didn't exist, add it
	node.Attributes = append(node.Attributes, attr, value)
}

func NodeRemoveAttribute(node *gcdapi.DOMNode, attr string) {
	attr = strings.ToLower(attr)
	if node.Attributes == nil {
		node.Attributes = make([]string, 0)
	}
	idx := 0
	for i := 0; i < len(node.Attributes); i += 2 {
		if strings.ToLower(node.Attributes[i]) == attr {
			idx = i
			break
		}
	}
	if idx != 0 {
		node.Attributes = node.Attributes[idx : idx+1]
	}
}
