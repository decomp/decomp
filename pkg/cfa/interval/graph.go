package interval

import (
	"sort"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfg"
	"github.com/rickypai/natsort"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// Graph is a control flow graph which records structuring information.
type Graph struct {
	// Underlying graph.
	cfa.Graph
}

// NewGraph returns a new control flow graph.
func NewGraph() *Graph {
	return &Graph{
		Graph: cfg.NewGraph(),
	}
}

// NewNode returns a new Node with a unique arbitrary ID.
func (g *Graph) NewNode() graph.Node {
	return &Node{
		Node: g.Graph.NewNode().(cfa.Node),
	}
}

// NodesOf returns it.Len() nodes from it.
func NodesOf(nodes graph.Nodes) []*Node {
	var ns []*Node
	for nodes.Next() {
		n := nodes.Node().(*Node)
		ns = append(ns, n)
	}
	return ns
}

// Node is a control flow graph node.
type Node struct {
	// Underlying node.
	cfa.Node

	// Pre-order DFS visit number.
	PreNum int
	// Reverse post-order DFS visit number.
	RevPostNum int

	// Loop structuring information.

	// Loop header node; or nil if not part of loop.
	LoopHead *Node
	// Type of the loop.
	LoopType LoopType
	// Follow node of the loop.
	LoopFollow *Node

	// 2-way conditional structuring information.

	// Follow node of 2-way and n-way conditionals.
	Follow *Node

	// Compound conditional structuring information.

	// Node contains only conditional and branching information, no other
	// instructions.
	IsCondNode bool
	// Compound conditional used to represent short-circuit evaluation in the
	// control flow graph.
	// TODO: figure out a better representation; perhaps struct with two fields,
	// merged condnode name and boolean operation.
	//    NOT n AND t
	//    n OR t
	//    n AND e
	//    NOT n OR e
	CompCond string
}

// SetAttribute implements encoding.AttributeSetter for Node.
func (n *Node) SetAttribute(attr encoding.Attribute) error {
	switch attr.Key {
	case "condnode":
		n.IsCondNode = attr.Value == "true"
	}
	return n.Node.SetAttribute(attr)
}

// initDFSOrder initializes the DFS visit order of the control flow graph.
func initDFSOrder(g cfa.Graph) {
	preNum := 1
	pre := func(n *Node) {
		n.PreNum = preNum
		preNum++
	}
	n := g.Nodes().Len()
	revPostNum := n
	post := func(n *Node) {
		n.RevPostNum = revPostNum
		revPostNum--
	}
	DFS(g, pre, post)
}

// DFS performs a depth-first search of the control flow graph, invoking non-nil
// pre and post during pre- and post-order visit, respectively.
func DFS(g cfa.Graph, pre, post func(n *Node)) {
	visited := make(map[int64]bool)
	var dfs func(n *Node)
	dfs = func(n *Node) {
		if visited[n.ID()] {
			return
		}
		visited[n.ID()] = true
		if pre != nil {
			pre(n)
		}
		for _, succ := range successors(g, n.ID()) {
			dfs(succ)
		}
		if post != nil {
			post(n)
		}
	}
	dfs(g.Entry().(*Node))
}

// successors returns the immediate successors of the node with the given ID in
// the control flow graph. The successors are ordered based on the condition of
// their outgoing edge; true before false in 2-way conditionals, and case 1
// through case n before default in n-way conditionals.
func successors(g cfa.Graph, id int64) []*Node {
	nodes := NodesOf(g.From(id))
	less := func(i, j int) bool {
		ni := nodes[i]
		nj := nodes[j]
		ei := g.Edge(id, ni.ID()).(cfa.Edge)
		ej := g.Edge(id, nj.ID()).(cfa.Edge)
		ci, ok := ei.Attribute("cond")
		if !ok {
			// Fall-back to sorting on DOTID, to make output deterministic.
			return natsort.Less(ni.DOTID(), nj.DOTID())
		}
		cj, ok := ej.Attribute("cond")
		if !ok {
			// Fall-back to sorting on DOTID, to make output deterministic.
			return natsort.Less(ni.DOTID(), nj.DOTID())
		}
		switch {
		case ci == "true" && cj == "false":
			return true
		case ci == "false" && cj == "true":
			return false
		}
		// Figure out a better way to handle case conditions. Use natural sorting
		// order for now, as that sorts x == 1 before x == 2.
		return natsort.Less(ci, cj)
	}
	sort.Slice(nodes, less)
	return nodes
}

// descRevPostOrder returns the nodes in descending reverse post-order. In
// particular, the returned list contains innermost nodes before outmost nodes.
func descRevPostOrder(nodes []*Node) []*Node {
	less := func(i, j int) bool {
		// Place in descending order.
		return nodes[i].RevPostNum > nodes[j].RevPostNum
	}
	sort.Slice(nodes, less)
	return nodes
}

// ascRevPostOrder returns the nodes in ascending reverse post-order. In
// particular, the returned list contains outermost nodes before innermost
// nodes.
func ascRevPostOrder(nodes []*Node) []*Node {
	less := func(i, j int) bool {
		// Place in ascending order.
		return nodes[i].RevPostNum < nodes[j].RevPostNum
	}
	sort.Slice(nodes, less)
	return nodes
}
