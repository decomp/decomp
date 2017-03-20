// Package cfa implements control flow analysis of control flow graphs.
package cfa

import (
	"fmt"

	"github.com/decomp/decomp/graph/cfg"
	"github.com/gonum/graph"
)

// label returns the label of the node.
func label(n graph.Node) string {
	if n, ok := n.(*cfg.Node); ok {
		return n.Label
	}
	panic(fmt.Sprintf("invalid node type; expected *cfg.Node, got %T", n))
}
