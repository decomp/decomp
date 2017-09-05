// Package cfg provides access to control flow graphs of LLVM IR functions.
package cfg

import (
	"fmt"

	"github.com/graphism/simple"
	"github.com/llir/llvm/ir"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// Graph represents a control flow graph.
type Graph struct {
	*simple.DirectedGraph
	entry graph.Node
	// nodes maps from basic block label to graph node.
	nodes map[string]*Node
}

// New returns a new control flow graph based on the given function.
func New(f *ir.Function) *Graph {
	g := &Graph{
		DirectedGraph: simple.NewDirectedGraph(),
		nodes:         make(map[string]*Node),
	}
	// Force generate local IDs.
	_ = f.String()
	for i, block := range f.Blocks {
		from := g.NewNodeWithLabel(block.Name)
		if i == 0 {
			// Store entry node.
			g.entry = from
		}
		switch term := block.Term.(type) {
		case *ir.TermRet:
			// nothing to do.
		case *ir.TermBr:
			to := g.NewNodeWithLabel(term.Target.Name)
			g.NewEdgeWithLabel(from, to, "")
		case *ir.TermCondBr:
			t := g.NewNodeWithLabel(term.TargetTrue.Name)
			f := g.NewNodeWithLabel(term.TargetFalse.Name)
			g.NewEdgeWithLabel(from, t, "true")
			g.NewEdgeWithLabel(from, f, "false")
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
			panic(fmt.Errorf("support for terminator %T not yet implemented", term))
		}
	}
	return g
}

// Entry returns the entry node of the control flow graph.
func (g *Graph) Entry() graph.Node {
	return g.entry
}

// SetEntry sets the entry node of the control flow graph.
func (g *Graph) SetEntry(entry graph.Node) {
	g.entry = entry
}

// NodeByLabel returns the node with the given basic block label in the graph.
// The boolean return value indicates success.
func (g *Graph) NodeByLabel(label string) (*Node, bool) {
	n, ok := g.nodes[label]
	return n, ok
}

// SetNodeLabel sets the basic block label of the node.
func (g *Graph) SetNodeLabel(n graph.Node, label string) {
	nn, ok := n.(*Node)
	if !ok {
		panic(fmt.Errorf("invalid node type; expected *cfg.Node, got %T", n))
	}
	if _, ok := g.nodes[nn.Label]; !ok {
		panic(fmt.Errorf("unable to locate node with basic block label %q in graph", nn.Label))
	}
	if _, ok := g.nodes[label]; ok {
		panic(fmt.Errorf("basic block label %q already present in graph", label))
	}
	delete(g.nodes, nn.Label)
	nn.Label = label
	g.nodes[label] = nn
}

// RemoveNode removes n from the graph, as well as any edges attached to it. If
// the node is not in the graph it is a no-op.
func (g *Graph) RemoveNode(n graph.Node) {
	nn, ok := n.(*Node)
	if !ok {
		panic(fmt.Errorf("invalid node type; expected *cfg.Node, got %T", n))
	}
	delete(g.nodes, nn.Label)
	g.DirectedGraph.RemoveNode(n)
}

// Node represents a node of a control flow graph.
type Node struct {
	graph.Node
	// Basic block label.
	Label string
	// DOT attributes.
	Attrs
	// entry specifies if the node is the entry node of the control flow graph.
	entry bool
}

// NewNode returns a new graph node with a unique arbitrary ID.
func (g *Graph) NewNode() graph.Node {
	return &Node{
		Node:  g.DirectedGraph.NewNode(),
		Attrs: make(Attrs),
	}
}

// NewNodeWithLabel returns a new node with the given basic block label in the
// graph, or the existing node if already present.
func (g *Graph) NewNodeWithLabel(label string) *Node {
	if n, ok := g.nodes[label]; ok {
		return n
	}
	n := &Node{
		Node:  g.DirectedGraph.NewNode(),
		Label: label,
		Attrs: make(Attrs),
	}
	g.nodes[label] = n
	g.AddNode(n)
	return n
}

// DOTID returns the DOT node ID of the node.
func (n *Node) DOTID() string {
	return n.Label
}

// SetDOTID sets the DOT node ID of the node.
func (n *Node) SetDOTID(id string) {
	n.Label = id
}

// SetAttribute sets the attribute of the node.
func (n *Node) SetAttribute(attr encoding.Attribute) error {
	switch attr.Key {
	case "label":
		if attr.Value == "entry" {
			n.entry = true
		}
		n.Attrs[attr.Key] = attr.Value
	default:
		// ignore attribute.
	}
	return nil
}

// Edge represents an edge of a control flow graph.
type Edge struct {
	graph.Edge
	// Edge label.
	Label string
}

// NewEdge returns a new Edge from the source to the destination node.
func (g *Graph) NewEdge(from, to graph.Node) graph.Edge {
	return &Edge{
		Edge: g.DirectedGraph.NewEdge(from, to),
	}
}

// NewEdgeWithLabel returns a new edge from the source to the destination node
// in the graph, or the existing edge if already present.
func (g *Graph) NewEdgeWithLabel(from, to graph.Node, label string) *Edge {
	if e := g.Edge(from, to); e != nil {
		return e.(*Edge)
	}
	e := &Edge{
		Edge:  g.DirectedGraph.NewEdge(from, to),
		Label: label,
	}
	g.SetEdge(e)
	return e
}

// Attributes returns the attributes of the edge.
func (e *Edge) Attributes() []encoding.Attribute {
	if len(e.Label) > 0 {
		return []encoding.Attribute{{Key: "label", Value: e.Label}}
	}
	return nil
}

// SetAttribute sets the attribute of the edge.
func (e *Edge) SetAttribute(attr encoding.Attribute) error {
	switch attr.Key {
	case "label":
		e.Label = attr.Value
	default:
		// ignore attribute.
	}
	return nil
}
