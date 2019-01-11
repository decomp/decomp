package interval

import (
	"github.com/mewmew/lnp/pkg/cfa"
)

// intervals returns the intervals of the given control flow graph.
//
// Pre: G is a control flow graph.
// Post: the intervals of G are contained in the list Is.
//
// ref: Figure 6-8; Cifuentes' Reverse Comilation Techniques.
func intervals(g cfa.Graph) []*Interval {
	// Is := {}
	var Is []*Interval
	// H := {h}
	H := newQueue()
	entry := g.Entry().(*Node)
	H.push(entry)
	// for (all unprocessed n \in H) do
	for !H.empty() {
		// I(n) := {n}
		n := H.pop()
		I := NewInterval(g, n)
		// repeat
		for {
			// I(n) := I(n) + {m \in N | \forall p \in immedPred(m), p \in I(n)}
			added := false
			for nodes := g.Nodes(); nodes.Next(); {
				m := nodes.Node().(*Node)
				if m.ID() == entry.ID() {
					// skip entry node.
					continue
				}
				allPredsInI := true
				for preds := g.To(m.ID()); preds.Next(); {
					p := preds.Node()
					if I.Node(p.ID()) == nil {
						allPredsInI = false
						break
					}
				}
				if allPredsInI {
					added = true
					I.addNode(m)
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
			if m.ID() == entry.ID() {
				// skip entry node.
				continue
			}
			if H.has(m.ID()) {
				// skip nodes in queue.
				continue
			}
			if I.Node(m.ID()) != nil {
				// skip nodes in interval.
				continue
			}
			// keep node if it has a predecessor in the interval.
			hasPredInI := false
			for preds := g.To(m.ID()); preds.Next(); {
				p := preds.Node().(*Node)
				if I.Node(p.ID()) != nil {
					hasPredInI = true
					break
				}
			}
			if hasPredInI {
				H.push(m)
			}
		}
		// Is := Is + I(n)
		Is = append(Is, I)
	}
	return Is
}

// --- [ Queue ] ---------------------------------------------------------------

// A queue is a FIFO queue of nodes.
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

// push appends the node to the end of the queue, if not already present.
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

// has reports whether the node is present in the queue.
func (q *queue) has(id int64) bool {
	for _, n := range q.ns {
		if id == n.ID() {
			return true
		}
	}
	return false
}
