package interval

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"github.com/mewmew/lnp/pkg/cfg"
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
func structCompCond(g cfa.Graph, before, after func(g cfa.Graph, prim *primitive.Primitive)) []*primitive.Primitive {
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
		fmt.Println("=== [ compcond ] ===")
		for _, n := range ascRevPostOrder(NodesOf(g.Nodes())) {
			// Order of nSuccs matter, as we have nSuccs[0] denote the true branch
			// and nSuccs[1] denote the false branch.
			nSuccs := successors(g, n.ID())
			// if (nodeType(n) == 2-way)
			if len(nSuccs) == 2 {
				fmt.Println("n:", n.DOTID()) // TODO: remove debug output
				// t = succ[n, 1]
				t := nSuccs[0]
				fmt.Println("   t:", t.DOTID()) // TODO: remove debug output
				tSuccs := successors(g, t.ID()) // used to make output deterministic.
				// e = succ[n, 2]
				e := nSuccs[1]
				fmt.Println("   e:", e.DOTID()) // TODO: remove debug output
				eSuccs := successors(g, e.ID()) // used to make output deterministic.
				switch {
				// if ((nodeType(t) == 2-way) \land (numInst(t) == 1) \land (numInEdges(t) == 1))
				case len(tSuccs) == 2 && t.IsCondNode && g.To(t.ID()).Len() == 1:
					// if (succ[t, 1] == e)
					switch {
					case tSuccs[0].ID() == e.ID():
						// if (n && !t)
						// modifyGraph(\lnot n \land t)
						// TODO: figure out how to represent compound condition.
						// Wrong in Cifuentes', which states NOT n AND t. Should be n AND NOT t.
						compCond := fmt.Sprintf("%q AND NOT %q", n.DOTID(), t.DOTID())
						n.CompCond = compCond
						prim := modifyGraph(g, n, t, tSuccs[1], "comp_cond_a_AND_NOT_b", before, after)
						prims = append(prims, prim)
						// change = True
						change = true
						continue loop
					// else if (succ[t, 2] == e)
					case tSuccs[1].ID() == e.ID():
						// if (n && t)
						// modifyGraph(n \lor t)
						// TODO: figure out how to represent compound condition.
						// Wrong in Cifuentes', which states n OR t. Should be n AND t.
						compCond := fmt.Sprintf("%q AND %q", n.DOTID(), t.DOTID())
						n.CompCond = compCond
						prim := modifyGraph(g, n, t, tSuccs[0], "comp_cond_a_AND_b", before, after)
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
						// if (n || e)
						// modifyGraph(n \land e)
						// TODO: figure out how to represent compound condition.
						// Wrong in Cifuentes', which states n AND e. Should be n OR e.
						compCond := fmt.Sprintf("%q OR %q", n.DOTID(), e.DOTID())
						n.CompCond = compCond
						prim := modifyGraph(g, n, e, eSuccs[1], "comp_cond_a_OR_b", before, after)
						prims = append(prims, prim)
						// change = True
						change = true
						continue loop
					// else if (succ[e, 2] = t)
					case eSuccs[1].ID() == t.ID():
						// modifyGraph(\lnot n \lor e)
						// TODO: figure out how to represent compound condition.
						// Wrong in Cifuentes', which states NOT n OR e. Should be n OR NOT e.
						compCond := fmt.Sprintf("%q OR NOT %q", n.DOTID(), e.DOTID())
						n.CompCond = compCond
						prim := modifyGraph(g, n, e, eSuccs[0], "comp_cond_a_OR_NOT_b", before, after)
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

// modifyGraph modifies the control flow graph to merge the compound condition
func modifyGraph(g cfa.Graph, n, c, follow cfa.Node, compCond string, before, after func(g cfa.Graph, prim *primitive.Primitive)) *primitive.Primitive {
	// Create primitive.
	prim := &primitive.Primitive{
		Prim:  compCond,
		Entry: n.DOTID(),
		Nodes: map[string]string{
			"a": n.DOTID(),
			"b": c.DOTID(),
			//"follow": follow.DOTID(),
		},
	}
	if before != nil {
		before(g, prim)
	}
	// Merge n and c nodes.
	olde := g.Edge(n.ID(), c.ID()).(*cfg.Edge)
	attrs := olde.Attrs
	g.RemoveNode(c.ID())
	newe := g.NewEdge(n, follow).(*cfg.Edge)
	newe.Attrs = attrs
	g.SetEdge(newe)
	if after != nil {
		after(g, prim)
	}
	return prim
}
