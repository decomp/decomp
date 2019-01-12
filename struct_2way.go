package interval

import "github.com/mewmew/lnp/pkg/cfa"

// struct2way structures 2-way conditionals in the given control flow graph and
// its dominator tree.
//
// Pre: G is a control flow graph.
//
// Post: 2-way conditionals are marked in G. The follow node for all 2-way
//       conditionals is determined.
//
// ref: Figure 6-31; Cifuentes' Reverse Comilation Techniques.
func struct2way(g cfa.Graph, dom cfa.DominatorTree) {
	// unresoved := {}
	unresolved := newStack()
	// for (all nodes m in descending order)
	for _, m := range descRevPostOrder(NodesOf(g.Nodes())) {
		// if ((nodeType(m) == 2-way) \land (inHeadLatch(m) == False))
		succs := g.From(m.ID())
		// TODO: verify what is meant by inHeadLatch. Does this correspond to LoopHead?
		if succs.Len() == 2 && m.LoopHead == nil {
			// if (\exists n, n = max{i | immedDom(i) = m \land #inEdges(i) >= 2})
			var follow *Node
			for _, i := range dom.DominatedBy(m.ID()) {
				ii := i.(*Node)
				if g.To(ii.ID()).Len() < 2 {
					// Follow node has at least 2 in-edges.
					continue
				}
				if follow == nil || follow.RevPostNum < ii.RevPostNum {
					follow = ii
				}
			}
			if follow != nil {
				// follow(m) = n
				m.Follow = follow
				// for (all x \in unresolved)
				for !unresolved.empty() {
					x := unresolved.pop()
					// follow(x) = n
					x.Follow = follow
					//unresolved = unresolved - {x}
				}
			} else {
				// unresolved = unresolved \union {m}
				unresolved.push(m)
			}
		}
	}
}

// stack is a LIFO stack of nodes.
type stack struct {
	// List of nodes in stack.
	ns []*Node
}

// newStack returns a new LIFO stack of nodes.
func newStack() *stack {
	return &stack{
		ns: make([]*Node, 0),
	}
}

// push appends the node to the end of the stack.
func (s *stack) push(n *Node) {
	s.ns = append(s.ns, n)
}

// pop pops and returns the last node of the stack.
func (s *stack) pop() *Node {
	if s.empty() {
		panic("invalid call to pop; empty stack")
	}
	length := len(s.ns)
	n := s.ns[length-1]
	s.ns = s.ns[:length-1]
	return n
}

// empty reports whether the stack is empty.
func (s *stack) empty() bool {
	return len(s.ns) == 0
}
