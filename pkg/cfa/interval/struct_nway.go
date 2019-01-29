package interval

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"gonum.org/v1/gonum/graph"
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
				// Handle unstructured n-way subgraph with abnormal entry, ref
				// Figure 6-36 in Cifuentes'; as recognized when a successor of the
				// n-way conditional header node is not immediately dominated by the
				// header node.
				if dom.DominatorOf(s.ID()).ID() != m.ID() {
					// n = commonImmedDom({s | s = succ(m)})
					n = commonImmedDom(mSuccs, dom).(*Node)
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
			//
			// Candidate follow nodes are all nodes that have the header node as
			// immediate dominator, and that are not successors of this node. The
			// candidate follow node with the most paths from the header node that
			// reach it is considered the follow node of the complete subgraph.
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

// commonImmedDom returns the common immediate dominator of the given nodes.
func commonImmedDom(nodes []*Node, dom cfa.DominatorTree) graph.Node {
	// Copy ns slice, so we may alter its elements without effecting the caller.
	ns := make([]graph.Node, len(nodes))
	for i, n := range nodes {
		ns[i] = n
	}
	// Calculate distance from the root node of the control flow graph.
	dists := make([]int, len(ns))
	min := -1
	for i, n := range ns {
		dist := distFromRoot(n, dom)
		dists[i] = dist
		if min == -1 || dist < min {
			min = dist
		}
	}
	// Walk up the immediate dominators until all nodes are at the same distance
	// in number of immediate dominators from the root node of the control flow
	// graph.
	for i, n := range ns {
		if dists[i] != min {
			dists[i]--
			ns[i] = dom.DominatorOf(n.ID())
		}
	}
	// Walk up the immediate dominators, one level at the time, until all nodes
	// have a shared common dominator.
	for i := min; i > 0; i-- {
		common := true
		var cidom graph.Node
		for i, n := range ns {
			idom := dom.DominatorOf(n.ID())
			if cidom == nil {
				cidom = idom
			}
			ns[i] = idom
			if cidom != nil && idom.ID() != cidom.ID() {
				common = false
			}
		}
		if common {
			return cidom
		}
	}
	// Only the entry node has no dominators. Thus it must be the common
	// immediate dominator.
	return ns[0]
}

// distFromRoot returns the distance in number of immediate dominators from the
// given node to the root node of the control flow graph.
func distFromRoot(n graph.Node, dom cfa.DominatorTree) int {
	dist := 0
	for {
		idom := dom.DominatorOf(n.ID())
		if idom == nil {
			return dist
		}
		n = idom
		dist++
	}
}
