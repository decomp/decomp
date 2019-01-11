package interval

import (
	"fmt"
	"sort"

	"github.com/mewmew/lnp/pkg/cfa"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
)

type Interval struct {
	g     cfa.Graph
	head  *Node
	nodes map[int64]*Node
}

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
		panic(fmt.Errorf("node with ID %d already present in interval; prev `%v`, new `%v`", n.ID(), prev, n))
	}
	i.nodes[n.ID()] = n
}

// Node returns the node with the given ID if it exists in the interval, and nil
// otherwise.
func (i *Interval) Node(id int64) graph.Node {
	return i.nodes[id]
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
		return nk.DOTID() < nl.DOTID()
	}
	sort.Slice(nodes, less)
	return iterator.NewOrderedNodes(nodes)
}

// --- [ skip? ] ---
//
// TODO: skip these methods by embedding graph.Directed in interval and
// implementing only the Node and Nodes methods.

// From returns all nodes that can be reached directly from the node with the
// given ID.
//
// From must not return nil.
func (i *Interval) From(id int64) graph.Nodes {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.From(id)
}

// HasEdgeBetween returns whether an edge exists between nodes with IDs xid and
// yid without considering direction.
func (i *Interval) HasEdgeBetween(xid, yid int64) bool {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.HasEdgeBetween(xid, yid)
}

// Edge returns the edge from u to v, with IDs uid and vid, if such an edge
// exists and nil otherwise. The node v must be directly reachable from u as
// defined by the From method.
func (i *Interval) Edge(uid, vid int64) graph.Edge {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.Edge(uid, vid)
}

// HasEdgeFromTo returns whether an edge exists in the graph from u to v with
// IDs uid and vid.
func (i *Interval) HasEdgeFromTo(uid, vid int64) bool {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.HasEdgeFromTo(uid, vid)
}

// To returns all nodes that can reach directly to the node with the given ID.
//
// To must not return nil.
func (i *Interval) To(id int64) graph.Nodes {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.To(id)
}

// --- [/ skip? ] ---
