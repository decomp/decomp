// Package cfg declares the types used to represent control flow graphs.
package cfg

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/graphism/simple"
	"github.com/mewmew/lnp/cfa"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// Graph is a control flow graph rooted at entry.
type Graph struct {
	// Entry node of control flow graph.
	entry *Node
	// Underlying simple.DirectedGraph.
	*simple.DirectedGraph
	// DOT graph ID.
	dotID string
	// nodes maps from DOT node ID to associated node.
	nodes map[string]*Node
}

// ParseFile parses the given Graphviz DOT file into a control flow graph.
func ParseFile(dotPath string) (*Graph, error) {
	// Parse DOT file.
	data, err := ioutil.ReadFile(dotPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	g := &Graph{
		DirectedGraph: simple.NewDirectedGraph(),
		nodes:         make(map[string]*Node),
	}
	if err := dot.Unmarshal(data, g); err != nil {
		return nil, errors.WithStack(err)
	}
	// Locate entry node.
	nodes := g.Nodes()
	for nodes.Next() {
		// Note: This run-time type assertion goes away, should Gonum graph start
		// to leverage generics in Go2.
		n := nodes.Node().(*Node)
		if _, entry := n.Attrs["entry"]; entry {
			if g.entry != nil {
				return nil, errors.Errorf("multiple entry nodes in control flow graph; prev %q, new %q", g.entry.DOTID(), n.DOTID())
			}
			g.entry = n
		}
	}
	if g.entry == nil {
		return nil, errors.Errorf("unable to locate entry node of control flow graph %q", g.DOTID())
	}
	return g, nil
}

// NewNode returns a new Node with a unique arbitrary ID.
func (g *Graph) NewNode() graph.Node {
	return &Node{
		Node:  g.DirectedGraph.NewNode(),
		Attrs: make(Attrs),
	}
}

// NewEdge returns a new Edge from the source to the destination node.
func (g *Graph) NewEdge(from, to graph.Node) graph.Edge {
	return &Edge{
		Edge:  g.DirectedGraph.NewEdge(from, to),
		Attrs: make(Attrs),
	}
}

// DOTID implements the dot.Graph interface for Graph.
func (g *Graph) DOTID() string {
	return g.dotID
}

// SetDOTID implements the dot.DOTIDSetter interface for Graph.
func (g *Graph) SetDOTID(dotID string) {
	g.dotID = dotID
}

// Entry returns the entry node of the control flow graph.
func (g *Graph) Entry() cfa.Node {
	return g.entry
}

// SetEntry sets the entry node of the control flow graph to entry.
func (g *Graph) SetEntry(entry cfa.Node) {
	// Note: This run-time type assertion goes away, should Gonum graph start to
	// leverage generics in Go2.
	g.entry = entry.(*Node)
}

// NodeWithDOTID returns the node with the given DOT node ID in the control flow
// graph. The boolean return value indicates success.
func (g *Graph) NodeWithDOTID(dotID string) (cfa.Node, bool) {
	n, ok := g.nodes[dotID]
	return n, ok
}

// AddNode adds a node to the graph. AddNode panics if the added node ID matches
// an existing node ID.
func (g *Graph) AddNode(n graph.Node) {
	// Note: This run-time type assertion goes away, should Gonum graph start to
	// leverage generics in Go2.
	nn := n.(*Node)
	dotID := nn.DOTID()
	if prev, ok := g.nodes[dotID]; ok {
		panic(fmt.Errorf("node with DOT node ID %q already present; prev `%v`, new `%v`", dotID, prev, nn))
	}
	g.nodes[dotID] = nn
	// Update entry node.
	if _, ok := nn.Attrs["entry"]; ok {
		g.entry = nn
	}
	g.DirectedGraph.AddNode(nn)
}

// RemoveNode removes the node with the given ID from the graph, as well as any
// edges attached to it. If the node is not in the graph it is a no-op.
func (g *Graph) RemoveNode(id int64) {
	// Note: This run-time type assertion goes away, should Gonum graph start to
	// leverage generics in Go2.
	n := g.Node(id).(*Node)
	if _, ok := n.Attrs["entry"]; ok {
		// Remove entry node.
		g.entry = nil
	}
	dotID := n.DOTID()
	delete(g.nodes, dotID)
	g.DirectedGraph.RemoveNode(id)
}

// String returns the string representation of the control flow graph in
// Graphviz DOT format.
func (g *Graph) String() string {
	buf, err := dot.Marshal(g, "", "", "\t")
	if err != nil {
		panic(fmt.Errorf("unable to marshal control flow graph to DOT format; %v", err))
	}
	return string(buf)
}

// --- [ Node ] ----------------------------------------------------------------

// Node is a node of the control flow graph.
type Node struct {
	// Underlying simple.Node.
	graph.Node
	// Node attributes.
	Attrs
	// DOT node ID.
	dotID string
}

// DOTID implements the dot.Node interface for Node.
func (n *Node) DOTID() string {
	return n.dotID
}

// SetDOTID implements the dot.DOTIDSetter interface for Node.
func (n *Node) SetDOTID(dotID string) {
	n.dotID = dotID
}

// --- [ Edge ] ----------------------------------------------------------------

// Edge is an edge of the control flow graph.
type Edge struct {
	// Underlying simple.Edge.
	graph.Edge
	// Edge attributes.
	Attrs
}

// --- [ Attributes ] ----------------------------------------------------------

// Attrs is a set of key-value pair attributes used by graph.Node or graph.Edge.
type Attrs map[string]string

// Attributes implements encoding.Attributer for Attrs.
func (a Attrs) Attributes() []encoding.Attribute {
	attrs := make([]encoding.Attribute, 0, len(a))
	for key, val := range a {
		attr := encoding.Attribute{Key: key, Value: val}
		attrs = append(attrs, attr)
	}
	// Sort by key.
	less := func(i, j int) bool {
		return attrs[i].Key < attrs[j].Key
	}
	sort.Slice(attrs, less)
	return attrs
}

// SetAttribute implements encoding.AttributeSetter for Attrs.
func (a Attrs) SetAttribute(attr encoding.Attribute) error {
	a[attr.Key] = attr.Value
	return nil
}

// Attribute returns the value of the attribute with the given key. The boolean
// return value indicates success.
func (a Attrs) Attribute(key string) (string, bool) {
	val, ok := a[key]
	return val, ok
}

// DelAttribute deletes the attribute with the given key.
func (a Attrs) DelAttribute(key string) {
	delete(a, key)
}
