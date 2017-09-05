// Package cfa implements control flow analysis of control flow graphs.
package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
)

// FindPrim locates a control flow primitive in the provided control flow graph
// and merges its nodes into a single node.
func FindPrim(g graph.Directed, dom cfg.DominatorTree) (*primitive.Primitive, error) {
	// Locate pre-test loops.
	if prim, ok := FindPreLoop(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate post-test loops.
	if prim, ok := FindPostLoop(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate 1-way conditionals.
	if prim, ok := FindIf(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate 1-way conditionals with a body return statements.
	if prim, ok := FindIfReturn(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate 2-way conditionals.
	if prim, ok := FindIfElse(g, dom); ok {
		return prim.Prim(), nil
	}

	// TODO: Locate n-way conditionals.
	//if prim, ok := FindSwitch(g, dom); ok {
	//	return prim.Prim(), nil
	//}

	// Locate sequences of two statements.
	if prim, ok := FindSeq(g, dom); ok {
		return prim.Prim(), nil
	}

	return nil, errors.New("unable to locate control flow primitive")
}

// Merge merges the nodes of the primitive into a single node, which is assigned
// the basic block label of the entry node.
func Merge(g *cfg.Graph, prim *primitive.Primitive) error {
	// Locate nodes to merge.
	var nodes []graph.Node
	for _, label := range prim.Nodes {
		node, ok := g.NodeByLabel(label)
		if !ok {
			return errors.Errorf("unable to locate pre-merge node label %q", label)
		}
		nodes = append(nodes, node)
	}
	entry, ok := g.NodeByLabel(prim.Entry)
	if !ok {
		return errors.Errorf("unable to locate entry node label %q", prim.Entry)
	}
	exit, ok := g.NodeByLabel(prim.Exit)
	if !ok {
		return errors.Errorf("unable to locate exit node label %q", prim.Exit)
	}

	// Add new node for primitive.
	entryLabel := entry.Label
	p := g.NewNodeWithLabel(fmt.Sprintf("prim_node_of_%s", entryLabel))

	// Connect incoming edges to entry.
	for _, from := range g.To(entry) {
		e := g.Edge(from, entry)
		var label string
		if e, ok := e.(*cfg.Edge); ok {
			label = e.Label
		}
		g.NewEdgeWithLabel(from, p, label)
	}

	// Connect outgoing edges from exit.
	for _, to := range g.From(exit) {
		e := g.Edge(exit, to)
		var label string
		if e, ok := e.(*cfg.Edge); ok {
			label = e.Label
		}
		g.NewEdgeWithLabel(p, to, label)
	}

	// Remove old nodes.
	for _, node := range nodes {
		g.RemoveNode(node)
	}

	// Set label of new primitive to the label of the entry node.
	g.SetNodeLabel(p, entryLabel)
	g.SetEntry(p)

	return nil
}

// label returns the label of the node.
func label(n graph.Node) string {
	if n, ok := n.(*cfg.Node); ok {
		return n.Label
	}
	panic(fmt.Sprintf("invalid node type; expected *cfg.Node, got %T", n))
}
