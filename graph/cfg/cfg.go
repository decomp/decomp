// Package cfg provides access to control flow graphs of LLVM IR functions.
package cfg

import (
	"fmt"

	"github.com/graphism/simple"
	"github.com/llir/llvm/ir"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
)

// New returns a new control flow graph based on the given function.
func New(f *ir.Function) *Graph {
	// Force generate local IDs.
	_ = f.String()
	g := newGraph()
	for _, block := range f.Blocks {
		from := g.NewNodeWithLabel(block.Name)
		switch term := block.Term.(type) {
		case *ir.TermRet:
			// nothing to do.
		case *ir.TermBr:
			to := g.NewNodeWithLabel(term.Target.Name)
			g.NewEdgeWithLabel(from, to, "")
		case *ir.TermCondBr:
			to := g.NewNodeWithLabel(term.TargetTrue.Name)
			g.NewEdgeWithLabel(from, to, "true")
			to = g.NewNodeWithLabel(term.TargetFalse.Name)
			g.NewEdgeWithLabel(from, to, "false")
		case *ir.TermSwitch:
			for _, c := range term.Cases {
				to := g.NewNodeWithLabel(c.Target.Name)
				label := fmt.Sprintf("case (x=%v)", c.X.Ident())
				g.NewEdgeWithLabel(from, to, label)
			}
			to := g.NewNodeWithLabel(term.TargetDefault.Name)
			g.NewEdgeWithLabel(from, to, "default case")
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
	nodes map[string]*Node
}

// newGraph returns a new control flow graph.
func newGraph() *Graph {
	g := &Graph{
		DirectedGraph: simple.NewDirectedGraph(0, 0),
		nodes:         make(map[string]*Node),
	}
	return g
}

// RemoveNode removes n from the graph, as well as any edges attached to it. If
// the node is not in the graph it is a no-op.
func (g *Graph) RemoveNode(n graph.Node) {
	if n, ok := n.(*Node); ok {
		delete(g.nodes, n.Label)
	}
	g.DirectedGraph.RemoveNode(n)
}

// NodeByLabel returns the node in the graph with the given label.
func (g *Graph) NodeByLabel(label string) *Node {
	return g.nodes[label]
}

// newNode returns a new node with a unique node ID in the graph.
func (g *Graph) newNode() *Node {
	n := &Node{
		Node:  g.NewNode(),
		Attrs: make(Attrs),
	}
	g.AddNode(n)
	return n
}

// NewNodeWithLabel returns a new node with the given label and a unique node ID
// in the graph, or the existing edge if already present.
func (g *Graph) NewNodeWithLabel(label string) *Node {
	if n, ok := g.nodes[label]; ok {
		return n
	}
	n := g.newNode()
	n.Label = label
	if len(label) > 0 {
		n.Attrs["label"] = label
	}
	g.nodes[label] = n
	return n
}

// newEdge returns a new edge from the source to the destination node in the
// graph, or the existing edge if already present.
func (g *Graph) newEdge(from, to graph.Node) *Edge {
	if e := g.Edge(from, to); e != nil {
		return e.(*Edge)
	}
	e := &Edge{
		Edge: simple.Edge{
			F: from,
			T: to,
		},
		Attrs: make(Attrs),
	}
	g.SetEdge(e)
	return e
}

// NewEdgeWithLabel returns a new edge from the source to the destination node
// and with the given label in the graph, or the existing edge if already
// present.
func (g *Graph) NewEdgeWithLabel(from, to graph.Node, label string) *Edge {
	e := g.newEdge(from, to)
	e.Label = label
	if len(label) > 0 {
		e.Attrs["label"] = label
	}
	return e
}

// SetLabel sets the label of the given node.
func (g *Graph) SetLabel(n graph.Node, label string) error {
	if _, ok := g.nodes[label]; ok {
		return errors.Errorf("node %q already present in graph", label)
	}
	node, ok := n.(*Node)
	if !ok {
		return errors.Errorf("expected node type *cfa.Node, got %T", n)
	}
	node.Label = label
	g.nodes[label] = node
	return nil
}

// Node represents a node of a control flow graph.
type Node struct {
	graph.Node
	// Node label.
	Label string
	// DOT attributes.
	Attrs
}

// Edge represents an edge of a control flow graph.
type Edge struct {
	simple.Edge
	// Edge label.
	Label string
	// DOT attributes.
	Attrs
}
