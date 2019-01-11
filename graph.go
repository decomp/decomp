package interval

import (
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfg"
	"gonum.org/v1/gonum/graph"
)

type Graph struct {
	cfa.Graph
}

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

type Node struct {
	cfa.Node
}
