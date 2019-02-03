package interval

import (
	"fmt"
	"sort"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/rickypai/natsort"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/iterator"
)

// An Interval I(h) with header node h is a maximal single-entry subgraph -- of
// a control flow graph rooted at entry -- in which h is the only entry node and
// all cycles contain h.
type Interval struct {
	g     cfa.Graph
	head  *Node
	nodes map[int64]*Node
}

// NewInterval returns a new interval in the control flow graph based on the
// given header node.
func NewInterval(g cfa.Graph, head *Node) *Interval {
	return &Interval{
		g:    g,
		head: head,
		nodes: map[int64]*Node{
			head.ID(): head,
		},
	}
}

// addNode adds a node to the interval. addNode panics if the added node ID
// matches an existing node ID.
func (i *Interval) addNode(n *Node) {
	if prev, ok := i.nodes[n.ID()]; ok {
		panic(fmt.Errorf("node with ID %d already present in interval; prev DOTID %q, new DOTID %q", n.ID(), prev.DOTID(), n.DOTID()))
	}
	i.nodes[n.ID()] = n
}

// Node returns the node with the given ID if it exists in the interval, and nil
// otherwise.
func (i *Interval) Node(id int64) graph.Node {
	n, ok := i.nodes[id]
	if !ok {
		// Ensure that nil is returned if node with ID is not present.
		//
		// Otherwise it would be converted to an interface value of graph.Node
		// type with value nil.
		return nil
	}
	return n
}

// Nodes returns all the nodes in the graph.
//
// Nodes must not return nil.
func (i *Interval) Nodes() graph.Nodes {
	var nodes []graph.Node
	for _, n := range i.nodes {
		nodes = append(nodes, n)
	}
	// Make order deterministic by sorting on DOTID.
	less := func(k, l int) bool {
		nk := i.nodes[nodes[k].ID()]
		nl := i.nodes[nodes[l].ID()]
		return natsort.Less(nk.DOTID(), nl.DOTID())
	}
	sort.Slice(nodes, less)
	return iterator.NewOrderedNodes(nodes)
}

// String returns the string representation of the interval in Graphviz DOT
// format.
func (i *Interval) String() string {
	buf, err := dot.Marshal(i, "", "", "\t")
	if err != nil {
		panic(fmt.Errorf("unable to marshal interval to DOT format; %v", err))
	}
	return string(buf)
}

// From returns all nodes that can be reached directly from the node with the
// given ID.
//
// From must not return nil.
func (i *Interval) From(id int64) graph.Nodes {
	var nodes []graph.Node
	for succs := i.g.From(id); succs.Next(); {
		succ := succs.Node()
		if i.Node(succ.ID()) == nil {
			// Skip node if not present in interval.
			continue
		}
		nodes = append(nodes, succ)
	}
	return iterator.NewOrderedNodes(nodes)
}

// HasEdgeBetween returns whether an edge exists between nodes with IDs xid and
// yid without considering direction.
func (i *Interval) HasEdgeBetween(xid, yid int64) bool {
	if i.Node(xid) == nil || i.Node(yid) == nil {
		return false
	}
	return i.g.HasEdgeBetween(xid, yid)
}

// Edge returns the edge from u to v, with IDs uid and vid, if such an edge
// exists and nil otherwise. The node v must be directly reachable from u as
// defined by the From method.
func (i *Interval) Edge(uid, vid int64) graph.Edge {
	if !i.HasEdgeBetween(uid, vid) {
		return nil
	}
	return i.g.Edge(uid, vid)
}

// HasEdgeFromTo returns whether an edge exists in the graph from u to v with
// IDs uid and vid.
func (i *Interval) HasEdgeFromTo(uid, vid int64) bool {
	if !i.HasEdgeBetween(uid, vid) {
		return false
	}
	return i.g.HasEdgeFromTo(uid, vid)
}

// To returns all nodes that can reach directly to the node with the given ID.
//
// To must not return nil.
func (i *Interval) To(id int64) graph.Nodes {
	var nodes []graph.Node
	for preds := i.g.To(id); preds.Next(); {
		pred := preds.Node()
		if i.Node(pred.ID()) == nil {
			// Skip node if not present in interval.
			continue
		}
		nodes = append(nodes, pred)
	}
	return iterator.NewOrderedNodes(nodes)
}
