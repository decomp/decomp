package interval

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
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
func structCompCond(g cfa.Graph) []*primitive.Primitive {
	var prims []*primitive.Primitive
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
			//fmt.Println("n:", n.DOTID()) // TODO: remove debug output
			// Order of nSuccs matter, as we have nSuccs[0] denote the true branch
			// and nSuccs[1] denote the false branch.
			nSuccs := successors(g, n.ID())
			// if (nodeType(n) == 2-way)
			if len(nSuccs) == 2 {
				// t = succ[n, 1]
				t := nSuccs[0]
				//fmt.Println("   t:", t.DOTID()) // TODO: remove debug output
				tSuccs := successors(g, t.ID()) // used to make output deterministic.
				// e = succ[n, 2]
				e := nSuccs[1]
				//fmt.Println("   e:", e.DOTID()) // TODO: remove debug output
				eSuccs := successors(g, e.ID()) // used to make output deterministic.
				switch {
				// if ((nodeType(t) == 2-way) \land (numInst(t) == 1) \land (numInEdges(t) == 1))
				case len(tSuccs) == 2 && t.IsCondNode && g.To(t.ID()).Len() == 1:
					// if (succ[t, 1] == e)
					switch {
					case tSuccs[0].ID() == e.ID():
						// modifyGraph(\lnot n \land t)
						// TODO: figure out how to represent compound condition.
						compCond := fmt.Sprintf("NOT %q AND %q", n.DOTID(), t.DOTID())
						n.CompCond = compCond
						g.RemoveNode(t.ID())
						g.SetEdge(g.NewEdge(n, tSuccs[1]))
						// Create primitive.
						prim := &primitive.Primitive{
							Prim:  "comp_cond_NOT_a_AND_b",
							Entry: n.DOTID(),
							Nodes: map[string]string{
								"a": n.DOTID(),
								"b": t.DOTID(),
								// TODO: add e as follow?
							},
						}
						prims = append(prims, prim)
						// change = True
						change = true
						continue loop
					// else if (succ[t, 2] == e)
					case tSuccs[1].ID() == e.ID():
						// modifyGraph(n \lor t)
						// TODO: figure out how to represent compound condition.
						compCond := fmt.Sprintf("%q OR %q", n.DOTID(), t.DOTID())
						n.CompCond = compCond
						g.RemoveNode(t.ID())
						g.SetEdge(g.NewEdge(n, tSuccs[0]))
						// Create primitive.
						prim := &primitive.Primitive{
							Prim:  "comp_cond_a_OR_b",
							Entry: n.DOTID(),
							Nodes: map[string]string{
								"a": n.DOTID(),
								"b": t.DOTID(),
								// TODO: add e as follow?
							},
						}
						prims = append(prims, prim)
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
						compCond := fmt.Sprintf("%q AND %q", n.DOTID(), e.DOTID())
						n.CompCond = compCond
						g.RemoveNode(e.ID())
						g.SetEdge(g.NewEdge(n, eSuccs[1]))
						// Create primitive.
						prim := &primitive.Primitive{
							Prim:  "comp_cond_a_AND_b",
							Entry: n.DOTID(),
							Nodes: map[string]string{
								"a": n.DOTID(),
								"b": e.DOTID(),
								// TODO: add t as follow?
							},
						}
						prims = append(prims, prim)
						// change = True
						change = true
						continue loop
					// else if (succ[e, 2] = t)
					case eSuccs[1].ID() == t.ID():
						// modifyGraph(\lnot n \lor e)
						// TODO: figure out how to represent compound condition.
						compCond := fmt.Sprintf("NOT %q OR %q", n.DOTID(), e.DOTID())
						n.CompCond = compCond
						g.RemoveNode(e.ID())
						g.SetEdge(g.NewEdge(n, eSuccs[0]))
						// Create primitive.
						prim := &primitive.Primitive{
							Prim:  "comp_cond_NOT_a_OR_b",
							Entry: n.DOTID(),
							Nodes: map[string]string{
								"a": n.DOTID(),
								"b": e.DOTID(),
								// TODO: add t as follow?
							},
						}
						prims = append(prims, prim)
						// change = True
						change = true
						continue loop
					}
				}
			}
		}
	}
	return prims
}
