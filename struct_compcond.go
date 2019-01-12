package interval

import (
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
	for change {
		// change = False
		change = false
		// for (all nodes n in postorder)
		// TODO: determine if we should use desc or asc.
		for _, n := range descRevPostOrder(NodesOf(g.Nodes())) {
			// Order of nSuccs matter, as we have nSuccs[0] denote the true branch
			// and nSuccs[1] denote the false branch.
			nSuccs := successors(g, n.ID())
			// if (nodeType(n) == 2-way)
			if len(nSuccs) == 2 {
				// t = succ[n, 1]
				t := nSuccs[0]
				tSuccs := NodesOf(g.From(t.ID()))
				// e = succ[n, 2]
				e := nSuccs[1]
				eSuccs := NodesOf(g.From(e.ID()))
				switch {
				// if ((nodeType(t) == 2-way) \land (numInst(t) == 1) \land (numInEdges(t) == 1))
				case len(tSuccs) == 2 && t.IsCondNode && g.To(t.ID()).Len() == 1:
					// if (succ[t, 1] == e)
					switch {
					case tSuccs[0].ID() == e.ID():
						// modifyGraph(\lnot n \land t)
						// TODO: figure out how to modify graph.
						log.Println("figure out how to modify graph: NOT n AND t")
						// change = True
						change = true
					// else if (succ[t, 2] == e)
					case tSuccs[1].ID() == e.ID():
						// modifyGraph(n \lor t)
						// TODO: figure out how to modify graph.
						log.Println("figure out how to modify graph: n OR t")
						// change = True
						change = true
					}
				// else if ((nodeType(e) == 2-way) \land (numInst(e) == 1) \land (numInEdges(e) == 1))
				case len(eSuccs) == 2 && e.IsCondNode && g.To(e.ID()).Len() == 1:
					switch {
					// if (succ[e, 1] = t)
					case eSuccs[0].ID() == t.ID():
						// modifyGraph(n \land e)
						// TODO: figure out how to modify graph.
						log.Println("figure out how to modify graph: n AND e")
						// change = True
						change = true
					// else if (succ[e, 2] = t)
					case eSuccs[1].ID() == t.ID():
						// modifyGraph(\lnot n \lor e)
						// TODO: figure out how to modify graph.
						log.Println("figure out how to modify graph: NOT n OR e")
						// change = True
						change = true
					}
				}
			}
		}
	}
}
