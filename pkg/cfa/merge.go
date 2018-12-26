package cfa

import (
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/encoding"
)

// Merge merges the nodes of the primitive into a single node, which is
// assigned the basic block label of the entry node.
func Merge(g Graph, prim *primitive.Primitive) (Graph, error) {
	// Set of nodes marked for removal; indexed by DOT node ID.
	primNodes := make(map[string]bool)
	for _, dotID := range prim.Nodes {
		primNodes[dotID] = true
	}
	// Locate nodes to remove, and check if entry node is one of the nodes being
	// deleted.
	var removeIDs []int64
	entry := false
	for dotID := range primNodes {
		n, ok := g.NodeWithDOTID(dotID)
		if !ok {
			return nil, errors.Errorf("unable to locate node with DOT node ID %q in control flow graph %q", dotID, g.DOTID())
		}
		removeIDs = append(removeIDs, n.ID())
		if isEntry(n) {
			entry = true
		}
	}

	// Create new node.
	//
	// Note: This run-time type assertion goes away, should Gonum graph start to
	// leverage generics in Go2.
	newNode := g.NewNode().(Node)
	newNode.SetDOTID(prim.Entry)
	if entry {
		attr := encoding.Attribute{Key: "entry", Value: "true"}
		newNode.SetAttribute(attr)
	}

	// Connect incoming edges of nodes being deleted to new node.
	var newEdges []Edge
	for _, removeID := range removeIDs {
		for preds := g.To(removeID); preds.Next(); {
			// Note: This run-time type assertion goes away, should Gonum graph
			// start to leverage generics in Go2.
			pred := preds.Node().(Node)
			if primNodes[pred.DOTID()] {
				// Skip edges from nodes being deleted.
				continue
			}
			// Note: This run-time type assertion goes away, should Gonum graph
			// start to leverage generics in Go2.
			e := g.Edge(pred.ID(), removeID).(Edge)
			// Note: This run-time type assertion goes away, should Gonum graph
			// start to leverage generics in Go2.
			newEdge := g.NewEdge(pred, newNode).(Edge)
			for _, attr := range e.Attributes() {
				newEdge.SetAttribute(attr)
			}
			newEdges = append(newEdges, newEdge)
		}
	}

	// Connect outgoing edges of nodes being deleted to new node.
	for _, removeID := range removeIDs {
		for succs := g.From(removeID); succs.Next(); {
			// Note: This run-time type assertion goes away, should Gonum graph
			// start to leverage generics in Go2.
			succ := succs.Node().(Node)
			if primNodes[succ.DOTID()] {
				// Skip edges to nodes being deleted.
				continue
			}
			// Note: This run-time type assertion goes away, should Gonum graph
			// start to leverage generics in Go2.
			e := g.Edge(removeID, succ.ID()).(Edge)
			// Note: This run-time type assertion goes away, should Gonum graph
			// start to leverage generics in Go2.
			newEdge := g.NewEdge(newNode, succ).(Edge)
			for _, attr := range e.Attributes() {
				newEdge.SetAttribute(attr)
			}
			newEdges = append(newEdges, newEdge)
		}
	}

	// Remove nodes to be merged and their associated edges.
	for _, removeID := range removeIDs {
		g.RemoveNode(removeID)
	}

	// Add new node to graph.
	g.AddNode(newNode)
	// Add incoming and outgoing edges of new node to graph.
	for _, newEdge := range newEdges {
		g.SetEdge(newEdge)
	}

	return g, nil
}

// isEntry reports whether the given node is the entry node of a control flow
// graph.
func isEntry(n Node) bool {
	for _, attr := range n.Attributes() {
		if attr.Key == "entry" {
			return true
		}
	}
	return false
}
