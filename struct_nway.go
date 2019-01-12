package interval

import (
	"log"

	"github.com/mewmew/lnp/pkg/cfa"
)

// structNway structures n-way conditionals in the given control flow graph and
// its dominator tree.
//
// Pre: G is a control flow graph.
//
// Post: n-way conditionals are structured in G. The follow node is determined
//       for all n-way subgraphs.
//
// ref: Figure 6-37; Cifuentes' Reverse Comilation Techniques.
func structNway(g cfa.Graph, dom cfa.DominatorTree) {
	// unresoved := {}
	unresolved := newStack()
	// for (all nodes m in postorder)
	// TODO: determine the order to use.
	for _, m := range descRevPostOrder(NodesOf(g.Nodes())) {
		mSuccs := successors(g, m.ID())
		// if ((nodeType(m) == n-way))
		if len(mSuccs) > 2 {
			var n *Node
			// if (\exists s : succ(m), immedDom(s) != m)
			for _, s := range mSuccs {
				if dom.DominatorOf(s.ID()).ID() != m.ID() {
					// n = commonImmedDom({s | s = succ(m)})
					// TODO: figure out how to implement commonImmedDom.
					log.Println("commonImmedDom not yet implemented")
				}
			}
			// else
			if n == nil {
				// n = m
				n = m
			}
			var follow *Node
			followNPreds := -1
			// if (\exists j, #inEdges(j) = max{i | immedDom(i) = n \land #inEdges(i) >= 2, #inEdges(i)})
			for _, i := range dom.DominatedBy(n.ID()) {
				ii := i.(*Node)
				npreds := g.To(ii.ID()).Len()
				if npreds >= 2 {
					if npreds > followNPreds {
						follow = ii
						followNPreds = npreds
					}
				}
			}
			if follow != nil {
				// follow(m) = j
				m.Follow = follow
				// for (all i \in unresolved)
				for !unresolved.empty() {
					x := unresolved.pop()
					// follow(i) = j
					x.Follow = follow
					// unresolved = unresolved - {i}
				}
			} else {
				// unresolved  = unresolved \union {m}
				unresolved.push(m)
			}
		}
	}
}
