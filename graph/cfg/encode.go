package cfg

import (
	"fmt"

	"github.com/gonum/graph/encoding/dot"
)

// DOTAttributes returns the DOT attributes of the graph node.
func (n *Node) DOTAttributes() []dot.Attribute {
	var attrs []dot.Attribute
	if len(n.Label) > 0 {
		attrs = append(attrs, newLabel(n.Label))
	}
	return attrs
}

// DOTAttributes returns the DOT attributes of the graph edge.
func (e *Edge) DOTAttributes() []dot.Attribute {
	var attrs []dot.Attribute
	if len(e.Label) > 0 {
		attrs = append(attrs, newLabel(e.Label))
	}
	return attrs
}

// newLabel returns a new label DOT attribute.
func newLabel(label string) dot.Attribute {
	return dot.Attribute{
		Key:   "label",
		Value: fmt.Sprintf("%q", label),
	}
}
