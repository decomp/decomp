package interval

import (
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfg"
	"gonum.org/v1/gonum/graph"
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
	LoopType loopType
	// Follow node of the loop.
	LoopFollow *Node
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
		for succs := g.From(n.ID()); succs.Next(); {
			succ := succs.Node().(*Node)
			dfs(succ)
		}
		if post != nil {
			post(n)
		}
	}
	dfs(g.Entry().(*Node))
}
