package cfg

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// A DominatorTree represents a dominator tree.
type DominatorTree struct {
	path.DominatorTree
}

// NewDom returns a new dominator tree based on the given graph.
func NewDom(g graph.Directed, entry graph.Node) DominatorTree {
	dt := path.Dominators(entry, g)
	return DominatorTree{
		DominatorTree: dt,
	}
}

// Dominates reports whether A dominates B.
func (dt DominatorTree) Dominates(a, b graph.Node) bool {
	return a == dt.DominatorTree.DominatorOf(b)
}
