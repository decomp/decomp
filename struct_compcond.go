package interval

import (
	"fmt"
	"log"

	"github.com/mewmew/lnp/pkg/cfa"
)

// structCompCond structures compound conditionals in the given control flow
// graph.
//
// Pre: G is a control flow graph.
//
//      2-way, n-way, and loops have been structured in G.
//
// Post: compound conditionals are structured in G.
//
// ref: Figure 6-34; Cifuentes' Reverse Comilation Techniques.
func structCompCond(g cfa.Graph) {
	// change = True
	change := true
	// while (change)
loop:
	for change {
		// change = False
		change = false
		// The algorithm to structure compound conditionals makes use of a
		// traversal from top to bottom of the graph, as the first condition in a
		// compound conditional expression is higher up in the graph (i.e. it is
		// tested first)
		//
		// for (all nodes n in postorder)
		for _, n := range ascRevPostOrder(NodesOf(g.Nodes())) {
			fmt.Println("n:", n.DOTID())
			// Order of nSuccs matter, as we have nSuccs[0] denote the true branch
			// and nSuccs[1] denote the false branch.
			nSuccs := successors(g, n.ID())
			// if (nodeType(n) == 2-way)
			if len(nSuccs) == 2 {
				// t = succ[n, 1]
				t := nSuccs[0]
				fmt.Println("   t:", t.DOTID())
				tSuccs := successors(g, t.ID()) // used to make output deterministic.
				// e = succ[n, 2]
				e := nSuccs[1]
				fmt.Println("   e:", e.DOTID())
				eSuccs := successors(g, e.ID()) // used to make output deterministic.
				switch {
				// if ((nodeType(t) == 2-way) \land (numInst(t) == 1) \land (numInEdges(t) == 1))
				case len(tSuccs) == 2 && t.IsCondNode && g.To(t.ID()).Len() == 1:
					// if (succ[t, 1] == e)
					switch {
					case tSuccs[0].ID() == e.ID():
						// modifyGraph(\lnot n \land t)
						// TODO: figure out how to represent compound condition.
						log.Println("figure out how to represent compound condition: NOT n AND t")
						compCond := fmt.Sprintf("NOT %q AND %q", n.DOTID(), t.DOTID())
						n.CompCond = compCond
						g.RemoveNode(t.ID())
						// change = True
						change = true
						continue loop
					// else if (succ[t, 2] == e)
					case tSuccs[1].ID() == e.ID():
						// modifyGraph(n \lor t)
						// TODO: figure out how to represent compound condition.
						log.Println("figure out how to represent compound condition: n OR t")
						compCond := fmt.Sprintf("%q OR %q", n.DOTID(), t.DOTID())
						n.CompCond = compCond
						g.RemoveNode(t.ID())
						// change = True
						change = true
						continue loop
					}
				// else if ((nodeType(e) == 2-way) \land (numInst(e) == 1) \land (numInEdges(e) == 1))
				case len(eSuccs) == 2 && e.IsCondNode && g.To(e.ID()).Len() == 1:
					switch {
					// if (succ[e, 1] = t)
					case eSuccs[0].ID() == t.ID():
						// modifyGraph(n \land e)
						// TODO: figure out how to represent compound condition.
						log.Println("figure out how to represent compound condition: n AND e")
						compCond := fmt.Sprintf("%q AND %q", n.DOTID(), e.DOTID())
						n.CompCond = compCond
						g.RemoveNode(e.ID())
						// change = True
						change = true
						continue loop
					// else if (succ[e, 2] = t)
					case eSuccs[1].ID() == t.ID():
						// modifyGraph(\lnot n \lor e)
						// TODO: figure out how to represent compound condition.
						log.Println("figure out how to represent compound condition: NOT n OR e")
						compCond := fmt.Sprintf("NOT %q OR %q", n.DOTID(), e.DOTID())
						n.CompCond = compCond
						g.RemoveNode(e.ID())
						// change = True
						change = true
						continue loop
					}
				}
			}
		}
	}
}
