// Package cfa implements control flow analysis of control flow graphs.
package cfa

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// Graph is a control flow graph and implements the graph.Directed,
// graph.Builder, graph.NodeRemover, graph.EdgeRemover and dot.Graph interfaces.
type Graph interface {
	// Entry returns the entry node of the control flow graph.
	Entry() Node
	// SetEntry sets the entry node of the control flow graph to entry.
	SetEntry(entry Node)
	graph.Directed
	graph.Builder
	graph.NodeRemover
	graph.EdgeRemover
	// DOTID returns the DOT graph ID of the control flow graph.
	DOTID() string
	dot.DOTIDSetter
	// NodeWithDOTID returns the node with the given DOT node ID in the control
	// flow graph. The boolean return value indicates success.
	NodeWithDOTID(dotID string) (Node, bool)
	// String returns the string representation of the control flow graph in
	// Graphviz DOT format.
	fmt.Stringer
}

// Node is a node of a control flow graph and implements the graph.Node,
// dot.Node, dot.DOTIDSetter, encoding.Attributer and encoding.AttributeSetter
// interfaces.
type Node interface {
	graph.Node
	dot.Node
	dot.DOTIDSetter
	Attributes
}

// Edge is an edge of a control flow graph and implements the graph.Edge,
// encoding.Attributer and encoding.AttributeSetter interfaces.
type Edge interface {
	graph.Edge
	Attributes
}

// NodesOf returns it.Len() nodes from it. It is safe to pass a nil Nodes to
// NodesOf.
func NodesOf(it graph.Nodes) []Node {
	nodes := graph.NodesOf(it)
	ns := make([]Node, len(nodes))
	for i, node := range nodes {
		// Note: This run-time type assertion goes away, should Gonum graph start
		// to leverage generics in Go2. Indeed, this entire function goes away.
		ns[i] = node.(Node)
	}
	return ns
}

// Attributes is a set of key-value pair attributes used by graph.Node or
// graph.Edge.
type Attributes interface {
	encoding.Attributer
	encoding.AttributeSetter
	// Attribute returns the value of the attribute with the given key. The
	// boolean return value indicates success.
	Attribute(key string) (string, bool)
	// DelAttribute deletes the attribute with the given key.
	DelAttribute(key string)
}
