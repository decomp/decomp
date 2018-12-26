package cfa

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// DominatorTree is a dominator tree of a control flow graph.
type DominatorTree struct {
	path.DominatorTree
}

// NewDom returns a new dominator tree based on the given control flow graph.
func NewDom(g Graph) DominatorTree {
	tree := path.Dominators(g.Entry(), g)
	return DominatorTree{
		DominatorTree: tree,
	}
}

// Dominates reports whether A dominates B.
func (dom DominatorTree) Dominates(a, b graph.Node) bool {
	return a == dom.DominatorTree.DominatorOf(b)
}
