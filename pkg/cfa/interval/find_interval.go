package interval

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
)

// Intervals returns the unique set of intervals of the given control flow
// graph.
//
// Pre: G is a control flow graph.
//
// Post: the intervals of G are contained in the list Is.
//
// ref: Allen, Frances E., and John Cocke. "A program data flow analysis
// procedure." Communications of the ACM 19.3 (1976): 137. [1]
//
// [1] https://pdfs.semanticscholar.org/81b9/49a01506a09fcd7ec4faf28e2fa0ec63f1e0.pdf
func Intervals(g cfa.Graph) []*Interval {
	var Is []*Interval
	// From [1] Algorithm for Finding Intervals.
	//
	// 1. Establish a set H for header nodes and initialize it with n_0, the
	//    unique entry node for the graph.
	H := newQueue()
	H.push(g.Entry().(*Node))
	// 2. For h \in H find I(h) as follows:
	for !H.empty() {
		h := H.pop()
		// 2.1. Put h in I(h) as the first element of I(h).
		I := NewInterval(g, h)
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
func findNodeWithImmPredsInInterval(g cfa.Graph, I *Interval) (*Node, bool) {
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
		return n.(*Node), true
	}
	return nil, false
}

// findUnusedNodeWithImmPredInInterval locates a node in G not in I nor H, which
// has one or more immediate predecessors in I. The boolean return value
// indicates success.
func findUnusedNodeWithImmPredInInterval(g cfa.Graph, I *Interval, H *queue) (*Node, bool) {
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
				return n.(*Node), true
			}
		}
	}
	return nil, false
}

// --- [ Queue ] ---------------------------------------------------------------

// queue is a FIFO queue of nodes which keeps track of all nodes that has been
// in the queue.
type queue struct {
	// List of nodes in queue.
	ns []*Node
	// Current position in queue.
	i int
}

// newQueue returns a new FIFO queue of nodes.
func newQueue() *queue {
	return &queue{
		ns: make([]*Node, 0),
	}
}

// push appends the node to the end of the queue if it has not been in the queue
// before.
func (q *queue) push(n *Node) {
	if !q.has(n.ID()) {
		q.ns = append(q.ns, n)
	}
}

// pop pops and returns the first node of the queue.
func (q *queue) pop() *Node {
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

// has reports whether the given node is present in the queue or has been
// present before.
func (q *queue) has(id int64) bool {
	for _, n := range q.ns {
		if id == n.ID() {
			return true
		}
	}
	return false
}
