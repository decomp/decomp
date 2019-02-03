// Package cfg declares the types used to represent control flow graphs.
package cfg

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/graphism/simple"
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// Graph is a control flow graph rooted at the entry node.
type Graph struct {
	// Entry node of control flow graph.
	entry cfa.Node
	// Underlying simple.DirectedGraph.
	*simple.DirectedGraph
	// DOT graph ID.
	dotID string
	// nodes maps from DOT node ID to associated node.
	nodes map[string]cfa.Node
}

// NewGraph returns a new control flow graph.
func NewGraph() *Graph {
	return &Graph{
		DirectedGraph: simple.NewDirectedGraph(),
		nodes:         make(map[string]cfa.Node),
	}
}

// ParseFile parses the given Graphviz DOT file into a control flow graph.
func ParseFile(dotPath string) (*Graph, error) {
	dst := NewGraph()
	err := ParseFileInto(dotPath, dst)
	return dst, err
}

// ParseFileInto parses the given Graphviz DOT file into the control flow graph
// dst.
func ParseFileInto(dotPath string, dst cfa.Graph) error {
	data, err := ioutil.ReadFile(dotPath)
	if err != nil {
		return errors.WithStack(err)
	}
	return ParseBytesInto(data, dst)
}

// Parse parses the given Graphviz DOT file into a control flow graph, reading
// from r.
func Parse(r io.Reader) (*Graph, error) {
	dst := NewGraph()
	err := ParseInto(r, dst)
	return dst, err
}

// ParseInto parses the given Graphviz DOT file into the control flow graph dst,
// reading from r.
func ParseInto(r io.Reader, dst cfa.Graph) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return ParseBytesInto(data, dst)
}

// ParseBytes parses the given Graphviz DOT file into a control flow graph,
// reading from data.
func ParseBytes(data []byte) (*Graph, error) {
	dst := NewGraph()
	err := ParseBytesInto(data, dst)
	return dst, err
}

// ParseBytesInto parses the given Graphviz DOT file into the control flow graph
// dst, reading from data.
func ParseBytesInto(data []byte, dst cfa.Graph) error {
	if err := dot.Unmarshal(data, dst); err != nil {
		return errors.WithStack(err)
	}
	// Locate entry node.
	for nodes := dst.Nodes(); nodes.Next(); {
		n := nodes.Node().(cfa.Node)
		if _, ok := n.Attribute("entry"); ok {
			prev := dst.Entry()
			if prev != nil {
				return errors.Errorf("multiple entry nodes in control flow graph; prev %q, new %q", prev.DOTID(), n.DOTID())
			}
			dst.SetEntry(n)
		}
	}
	if dst.Entry() == nil {
		return errors.Errorf("unable to locate entry node of control flow graph %q", dst.DOTID())
	}
	return nil
}

// ParseString parses the given Graphviz DOT file into a control flow graph, reading
// from s.
func ParseString(s string) (*Graph, error) {
	return ParseBytes([]byte(s))
}

// ParseStringInto parses the given Graphviz DOT file into the control flow graph
// dst, reading from s.
func ParseStringInto(s string, dst cfa.Graph) error {
	return ParseBytesInto([]byte(s), dst)
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
	if g.entry == nil {
		// Ensure that nil is returned if g.entry is nil.
		//
		// Otherwise it would be converted to an interface value of cfa.Node type
		// with value nil.
		return nil
	}
	return g.entry
}

// SetEntry sets the entry node of the control flow graph to entry.
func (g *Graph) SetEntry(entry cfa.Node) {
	entry.SetAttribute(encoding.Attribute{Key: "entry", Value: "true"})
	g.entry = entry
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
	nn := n.(cfa.Node)
	dotID := nn.DOTID()
	if prev, ok := g.nodes[dotID]; ok {
		panic(fmt.Errorf("node with DOT node ID %q already present; prev `%v`, new `%v`", dotID, prev, nn))
	}
	g.nodes[dotID] = nn
	// Update entry node.
	if _, ok := nn.Attribute("entry"); ok {
		g.entry = nn
	}
	g.DirectedGraph.AddNode(nn)
}

// RemoveNode removes the node with the given ID from the graph, as well as any
// edges attached to it. If the node is not in the graph it is a no-op.
func (g *Graph) RemoveNode(id int64) {
	n := g.Node(id).(cfa.Node)
	if _, ok := n.Attribute("entry"); ok {
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
