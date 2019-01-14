package cfa

import (
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

// Dominates reports whether node x dominates y, with node IDs xid and yid.
func (dom DominatorTree) Dominates(xid, yid int64) bool {
	dominator := dom.DominatorTree.DominatorOf(yid)
	return dominator != nil && dominator.ID() == xid
}
