// Package cfa implements control flow analysis of control flow graphs.
package cfa

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// Graph is a control flow graph and implements the graph.Directed and dot.Graph
// interfaces.
type Graph interface {
	// Entry returns the entry node of the control flow graph.
	Entry() Node
	// Underlying directed graph.
	graph.Directed
	// DOTID returns the DOT graph ID of the control flow graph.
	DOTID()
}

// Node is a node of a control flow graph and implements the graph.Node,
// dot.Node, encoding.Attributer and encoding.AttributeSetter interfaces.
type Node interface {
	graph.Node
	dot.Node
	encoding.Attributer
	encoding.AttributeSetter
}

// Edge is an edge of a control flow graph and implements the graph.Edge,
// encoding.Attributer and encoding.AttributeSetter interfaces.
type Edge interface {
	graph.Edge
	encoding.Attributer
	encoding.AttributeSetter
}
