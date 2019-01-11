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
	preNum int
	// Post-order DFS visit number.
	postNum int

	// Loop structuring information.

	// Specifies whether the node is part of a loop primitive.
	inLoop bool
	// Type of the loop.
	loopType loopType
	// Follow node of the loop.
	loopFollow *Node
}

// initDFSOrder initializes the DFS visit order of the control flow graph.
func initDFSOrder(g cfa.Graph) {
	preNum := 0
	pre := func(n *Node) {
		n.preNum = preNum
		preNum++
	}
	postNum := 0
	post := func(n *Node) {
		n.postNum = postNum
		postNum++
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
		for succs := g.From(n.ID()); succs.Next(); {
			succ := succs.Node().(*Node)
			dfs(succ)
		}
	}
	dfs(g.Entry().(*Node))
}
