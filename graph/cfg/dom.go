package cfg

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/flow"
)

// A DominatorTree represents a dominator tree.
type DominatorTree struct {
	flow.DominatorTree
}

// NewDom returns a new dominator tree based on the given graph.
func NewDom(g graph.Directed, entry graph.Node) DominatorTree {
	dt := flow.Dominators(entry, g)
	return DominatorTree{
		DominatorTree: dt,
	}
}

// Dominates reports whether A dominates B.
func (dt DominatorTree) Dominates(a, b graph.Node) bool {
	return a.ID() == dt.DominatorTree.DominatorOf(b.ID()).ID()
}
