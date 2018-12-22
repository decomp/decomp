// Package cfg implements control flow graph.
package cfg

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/graphism/simple"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// Graph is a control flow graph rooted at entry.
type Graph struct {
	// Entry node of control flow graph.
	Entry *Node
	// Underlying simple.DirectedGraph.
	*simple.DirectedGraph
	// DOT graph ID.
	dotID string
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
	}
	if err := dot.Unmarshal(data, g); err != nil {
		return nil, errors.WithStack(err)
	}
	// Locate entry node.
	for nodes := g.Nodes(); nodes.Next(); {
		n := nodes.Node()
		if n, ok := n.(*Node); ok {
			if _, entry := n.Attrs["entry"]; entry {
				if g.Entry != nil {
					return nil, errors.Errorf("multiple entry nodes in control flow graph; prev %q, new %q", g.Entry.DOTID(), n.DOTID())
				}
				g.Entry = n
			}
		}
	}
	if g.Entry == nil {
		return nil, errors.Errorf("unable to locate entry node of control flow graph %q", g.DOTID())
	}
	return g, nil
}

// NewNode returns a new Node with a unique arbitrary ID.
func (g *Graph) NewNode() graph.Node {
	fmt.Println("new node")
	return &Node{
		Node:  g.DirectedGraph.NewNode(),
		Attrs: make(Attrs),
	}
}

// NewEdge returns a new Edge from the source to the destination node.
func (g *Graph) NewEdge(from, to graph.Node) graph.Edge {
	fmt.Println("new edge")
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

// Edge is an edge of the control flow graph.
type Edge struct {
	// Underlying simple.Edge.
	graph.Edge
	// Edge attributes.
	Attrs
}

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
