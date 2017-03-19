// Package cfg provides access to control flow graphs of LLVM IR functions.
package cfg

import (
	"fmt"

	"github.com/gonum/graph"
	"github.com/gonum/graph/simple"
	"github.com/llir/llvm/ir"
)

// New returns a new control flow graph based on the given function.
func New(f *ir.Function) *Graph {
	// Force generate local IDs.
	_ = f.String()
	g := newGraph()
	for _, block := range f.Blocks {
		from := g.getNode(block.Name)
		switch term := block.Term.(type) {
		case *ir.TermRet:
			// nothing to do.
		case *ir.TermBr:
			to := g.getNode(term.Target.Name)
			g.setEdge(from, to, "")
		case *ir.TermCondBr:
			to := g.getNode(term.TargetTrue.Name)
			g.setEdge(from, to, "true")
			to = g.getNode(term.TargetFalse.Name)
			g.setEdge(from, to, "false")
		case *ir.TermSwitch:
			for _, c := range term.Cases {
				to := g.getNode(c.Target.Name)
				label := fmt.Sprintf("case (x=%v)", c.X.Ident())
				g.setEdge(from, to, label)
			}
			to := g.getNode(term.TargetDefault.Name)
			g.setEdge(from, to, "default case")
		case *ir.TermUnreachable:
			// nothing to do.
		default:
			panic(fmt.Sprintf("support for terminator %T not yet implemented", term))
		}
	}
	return g
}

// Graph represents a control flow graph, and implements the
// gonum/graph.DirectedBuilder interface.
type Graph struct {
	*simple.DirectedGraph
	// nodes maps from node label to graph node.
	nodes map[string]graph.Node
}

// newGraph returns a new control flow graph.
func newGraph() *Graph {
	g := &Graph{
		DirectedGraph: simple.NewDirectedGraph(0, 0),
		nodes:         make(map[string]graph.Node),
	}
	return g
}

// NodeByLabel returns the node in the graph with the given label.
func (g *Graph) NodeByLabel(label string) graph.Node {
	return g.nodes[label]
}

// getNode returns the node in the graph with the given label, generating a new
// such node if none exist.
func (g *Graph) getNode(label string) graph.Node {
	if n, ok := g.nodes[label]; ok {
		return n
	}
	id := g.NewNodeID()
	n := &Node{
		Node:  simple.Node(id),
		Label: label,
	}
	g.nodes[label] = n
	g.AddNode(n)
	return n
}

// setEdge adds an edge from the source to the destination node. An optional
// label may be specified for the edge.
func (g *Graph) setEdge(from, to graph.Node, label string) {
	e := &Edge{
		Edge: simple.Edge{
			F: from,
			T: to,
		},
		Label: label,
	}
	g.SetEdge(e)
}

// Node represents a node of a control flow graph.
type Node struct {
	simple.Node
	// Node label.
	Label string
}

// Edge represents an edge of a control flow graph.
type Edge struct {
	simple.Edge
	// Edge label.
	Label string
}
