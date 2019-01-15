package interval

import (
	"fmt"
	"log"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
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
func structNway(g cfa.Graph, dom cfa.DominatorTree) []*primitive.Primitive {
	var prims []*primitive.Primitive
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
				// Create primitive.
				prim := &primitive.Primitive{
					Prim:  "switch",
					Entry: m.DOTID(),
					Nodes: map[string]string{
						"cond":   m.DOTID(),
						"follow": follow.DOTID(),
					},
				}
				// follow(m) = j
				m.Follow = follow
				// for (all i \in unresolved)
				for i := 0; !unresolved.empty(); i++ {
					x := unresolved.pop()
					// follow(i) = j
					x.Follow = follow
					// unresolved = unresolved - {i}

					// Add loop body nodes to primitive.
					name := fmt.Sprintf("body_%d", i)
					prim.Nodes[name] = x.DOTID()
				}
				prims = append(prims, prim)
			} else {
				// unresolved  = unresolved \union {m}
				unresolved.push(m)
			}
		}
	}
	return prims
}
