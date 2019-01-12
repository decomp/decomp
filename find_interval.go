package interval

import (
	"github.com/mewmew/lnp/pkg/cfa"
)

// Intervals returns the unique set of intervals of the given control flow
// graph.
//
// Pre: G is a control flow graph.
//
// Post: the intervals of G are contained in the list Is.
//
// ref: Figure 6-8; Cifuentes' Reverse Comilation Techniques.
func Intervals(g cfa.Graph) []*Interval {
	// Is := {}
	var Is []*Interval
	// H := {h}
	H := newQueue()
	entry := g.Entry().(*Node)
	//fmt.Println("entry:", entry.DOTID()) // TODO: remove debug output
	H.push(entry)
	// for (all unprocessed n \in H) do
	for !H.empty() {
		// I(n) := {n}
		n := H.pop()
		//fmt.Println("==== n:", n.DOTID()) // TODO: remove debug output
		I := NewInterval(g, n)
		// repeat
		for {
			// I(n) := I(n) + {m \in N | \forall p \in immedPred(m), p \in I(n)}
			added := false
			for nodes := g.Nodes(); nodes.Next(); {
				m := nodes.Node().(*Node)
				if I.Node(m.ID()) != nil {
					// Interval already contains node.
					continue
				}
				//fmt.Println("m:", m.DOTID()) // TODO: remove debug output
				if m.ID() == entry.ID() {
					// skip entry node.
					continue
				}
				if containsAllPreds(g, I, m) {
					I.addNode(m)
					added = true
				}
			}
			// until no more nodes can be added to I(n).
			if !added {
				break
			}
		}
		// H := H + {m \in N | m \not \int H \land m \not in I(n) \land {\exists p \in immedPred(m), p \in I(n)}}
		for nodes := g.Nodes(); nodes.Next(); {
			m := nodes.Node().(*Node)
			//fmt.Printf("id: %v, dotid: %q\n", m.ID(), m.DOTID()) // TODO: remove debug output
			//fmt.Println("m2:", m.DOTID()) // TODO: remove debug output
			if m.ID() == entry.ID() {
				// skip entry node.
				continue
			}
			if H.has(m.ID()) {
				// skip nodes in queue.
				//fmt.Println("present in queue; skipping:", m.DOTID()) // TODO: remove debug output
				continue
			}
			if I.Node(m.ID()) != nil {
				// skip nodes in interval.
				//fmt.Println("m.ID():", m.ID()) // TODO: remove debug output
				//fmt.Println("already in interval; skipping:", m.DOTID()) // TODO: remove debug output
				continue
			}
			// keep node if it has a predecessor in the interval.
			hasPredInI := false
			for preds := g.To(m.ID()); preds.Next(); {
				p := preds.Node().(*Node)
				//fmt.Println("   p2:", p.DOTID()) // TODO: remove debug output
				if I.Node(p.ID()) != nil {
					hasPredInI = true
					break
				}
			}
			if hasPredInI {
				// Add node to queue.
				//fmt.Println("push to queue:", m.DOTID()) // TODO: remove debug output
				H.push(m)
			}
		}
		// Is := Is + I(n)
		Is = append(Is, I)
	}
	return Is
}

// containsAllPreds reports whether the interval I(h) contains all the immediate
// predecessors of n in the control flow graph, and whether n has at least one
// predecessor.
func containsAllPreds(g cfa.Graph, I *Interval, n cfa.Node) bool {
	preds := g.To(n.ID())
	if preds.Len() == 0 {
		// Ignore nodes without predecessors (e.g. entry node); otherwise they
		// would be added to every interval.
		return false
	}
	for preds.Next() {
		p := preds.Node()
		if I.Node(p.ID()) == nil {
			return false
		}
	}
	return true
}

// --- [ Queue ] ---------------------------------------------------------------

// A queue is a FIFO queue of nodes which keeps track of all nodes that has been
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
