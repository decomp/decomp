// Package interval implements the Interval method control flow recovery
// algorithm.
//
// At a high-level, the Interval method TODO...
package interval

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/mewkiz/pkg/term"
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
)

var (
	// dbg represents a logger with the "interval:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.YellowBold("interval:")+" ", 0)
	// warn represents a logger with the "interval:" prefix, which logs warning
	// messages to standard error.
	warn = log.New(os.Stderr, term.RedBold("interval:")+" ", 0)
)

// Analyze analyzes the given control flow graph and returns the list of
// recovered high-level control flow primitives. The before and after functions
// are invoked if non-nil before and after merging the nodes of located
// primitives.
func Analyze(g cfa.Graph, before, after func(g cfa.Graph, prim *primitive.Primitive)) ([]*primitive.Primitive, error) {
	panic("not yet implemented")
}

// --- [ Interval ] ------------------------------------------------------------

// ref: Allen, Frances E., and John Cocke. "A program data flow analysis
// procedure." Communications of the ACM 19.3 (1976): 137. [1]
//
// [1] https://pdfs.semanticscholar.org/81b9/49a01506a09fcd7ec4faf28e2fa0ec63f1e0.pdf

// intervals returns the intervals contained within the given control flow
// graph.
func intervals(g cfa.Graph) []*interval {
	var Is []*interval
	// From [1] Algorithm for Finding Intervals.
	//
	// 1. Establish a set H for header nodes and initialize it with n_0, the
	//    unique entry node for the graph.
	H := newQueue()
	H.push(g.Entry())
	// 2. For h \in H find I(h) as follows:
	for !H.empty() {
		h := H.pop()
		// 2.1. Put h in I(h) as the first element of I(h).
		I := newInterval(g, h)
		for {
			// 2.2. Add to I(h) any node all of whose immediate predecessors are
			//      already in I(h).
			n, ok := findNodeWithImmPredsInInterval(g, I)
			if !ok {
				// 2.3. Repeat 2.2 until no more nodes can be added to I(h).
				break
			}
			I.addNode(n)
		}
		// 3. Add to H all nodes in G which are not already in H and which are not
		//    in I(h) but which have immediate predecessors in I(h). Therefore a
		//    node is added to H the first time any (but not all) of its immediate
		//    predecessors become members of an interval.
		for {
			n, ok := findUnusedNodeWithImmPredInInterval(g, I, H)
			if !ok {
				break
			}
			H.push(n)
		}
		// 4. Add I(h) to a set Is of intervals being developed.
		Is = append(Is, I)
		// 5. Select the next unprocessed node in H and repeat steps 2, 3, 4, 5.
		//    When there are no more unprocessed nodes in H, the procedure
		//    terminates.
	}
	return Is
}

// findNodeWithImmPredsInInterval locates a node in G not in I, all of whose
// immediate predecessors in the interval I. The boolean return value indicates
// success.
func findNodeWithImmPredsInInterval(g cfa.Graph, I *interval) (cfa.Node, bool) {
	// 2.2. Add to I(h) any node all of whose immediate predecessors are already
	//      in I(h).

	// Note, we use rev post order here only to make output deterministic. The
	// computed interval would still be valid even without it.
loop:
	for _, n := range cfa.RevPostOrder(g) {
		if n.ID() == g.Entry().ID() {
			// skip entry node as it has no predecessors. Also, if entry is to be
			// part of I, then it is the header of I which is already added by
			// newInterval. constructor.
			continue
		}
		if I.Node(n.ID()) != nil {
			// skip if present in I.
			continue
		}
		preds := g.To(n.ID())
		// All nodes in a control flow graph, except the entry node, must have at
		// least one predecessor.
		if preds.Len() == 0 {
			panic(fmt.Errorf("invalid node %q; missing predecessors", n.DOTID()))
		}
		for preds.Next() {
			pred := preds.Node()
			if I.Node(pred.ID()) == nil {
				// skip node as not all immediate predecessors are in I.
				continue loop
			}
		}
		// node has all predecessors in I.
		return n, true
	}
	return nil, false
}

// findUnusedNodeWithImmPredInInterval locates a node in G not in I nor H, which
// has one or more immediate predecessors in I. The boolean return value
// indicates success.
func findUnusedNodeWithImmPredInInterval(g cfa.Graph, I *interval, H *queue) (cfa.Node, bool) {
	// 3. Add to H all nodes in G which are not already in H and which are not in
	//    I(h) but which have immediate predecessors in I(h). Therefore a node is
	//    added to H the first time any (but not all) of its immediate
	//    predecessors become members of an interval.

	// Note, we use rev post order here only to make output deterministic. The
	// computed interval would still be valid even without it.
	for _, n := range cfa.RevPostOrder(g) {
		if H.has(n.ID()) {
			// skip if present in H.
			continue
		}
		if I.Node(n.ID()) != nil {
			// skip if present in I.
			continue
		}
		preds := g.To(n.ID())
		// All nodes in a control flow graph, except the entry node, must have at
		// least one predecessor.
		if preds.Len() == 0 {
			panic(fmt.Errorf("invalid node %q; missing predecessors", n.DOTID()))
		}
		for preds.Next() {
			pred := preds.Node()
			if I.Node(pred.ID()) != nil {
				// node has at least one predecessor in I.
				return n, true
			}
		}
	}
	return nil, false
}

// An interval I(h) is the maximal, single-entry subgraph in which header node h
// is the only entry node and in which all closed paths contain h.
type interval struct {
	// Graph containing the interval.
	g cfa.Graph
	// head is the entry node of the interval.
	head cfa.Node
	// nodes tracks the nodes contained within the interval; mapping from node ID
	// to node.
	nodes map[int64]cfa.Node
}

// newInterval returns a new interval with the given header node.
func newInterval(g cfa.Graph, head cfa.Node) *interval {
	return &interval{
		g:    g,
		head: head,
		nodes: map[int64]cfa.Node{
			head.ID(): head,
		},
	}
}

// addNode adds a node to the interval. addNode panics if the added node ID
// matches an existing node ID.
func (i *interval) addNode(n cfa.Node) {
	if prev, ok := i.nodes[n.ID()]; ok {
		panic(fmt.Errorf("node with ID %d already present in interval; prev `%v`, new `%v`", n.ID(), prev, n))
	}
	i.nodes[n.ID()] = n
}

// Node returns the node with the given ID if it exists in the interval, and nil
// otherwise.
func (i *interval) Node(id int64) graph.Node {
	return i.nodes[id]
}

// Nodes returns all the nodes in the graph.
//
// Nodes must not return nil.
func (i *interval) Nodes() graph.Nodes {
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
func (i *interval) From(id int64) graph.Nodes {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.From(id)
}

// HasEdgeBetween returns whether an edge exists between nodes with IDs xid and
// yid without considering direction.
func (i *interval) HasEdgeBetween(xid, yid int64) bool {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.HasEdgeBetween(xid, yid)
}

// Edge returns the edge from u to v, with IDs uid and vid, if such an edge
// exists and nil otherwise. The node v must be directly reachable from u as
// defined by the From method.
func (i *interval) Edge(uid, vid int64) graph.Edge {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.Edge(uid, vid)
}

// HasEdgeFromTo returns whether an edge exists in the graph from u to v with
// IDs uid and vid.
func (i *interval) HasEdgeFromTo(uid, vid int64) bool {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.HasEdgeFromTo(uid, vid)
}

// To returns all nodes that can reach directly to the node with the given ID.
//
// To must not return nil.
func (i *interval) To(id int64) graph.Nodes {
	// TODO: determine if we only want to consider nodes in I, now we consider
	// nodes of the entire graph.
	return i.g.To(id)
}

// --- [/ skip? ] ---

// --- [ Queue ] ---------------------------------------------------------------

// A queue is a FIFO queue of nodes.
type queue struct {
	// List of nodes in queue.
	ns []cfa.Node
	// Current position in queue.
	i int
}

// newQueue returns a new FIFO queue of nodes.
func newQueue() *queue {
	return &queue{
		ns: make([]cfa.Node, 0),
	}
}

// push appends the node to the end of the queue, if not already present.
func (q *queue) push(n cfa.Node) {
	if !q.has(n.ID()) {
		q.ns = append(q.ns, n)
	}
}

// pop pops and returns the first node of the queue.
func (q *queue) pop() cfa.Node {
	if q.empty() {
		panic("invalid call to pop; empty queue")
	}
	n := q.ns[q.i]
	q.i++
	return n
}

// empty reports whether the queue is empty.
func (q *queue) empty() bool {
	return len(q.ns[q.i:]) == 0
}

// has reports whether the node is present in the queue.
func (q *queue) has(id int64) bool {
	for _, n := range q.ns {
		if id == n.ID() {
			return true
		}
	}
	return false
}
