package cfg

import "gonum.org/v1/gonum/graph"

// DFS performs a depth-first search of the control flow graph, starting at the
// entry node. The functions pre and post are invoked if non-nil during pre- and
// post-order traversal of the graph, respectively.
func DFS(g *Graph, pre, post func(n graph.Node)) {
	visited := make(map[graph.Node]bool)
	var visit func(n graph.Node)
	visit = func(n graph.Node) {
		visited[n] = true
		for succs := g.From(n.ID()); succs.Next(); {
			succ := succs.Node()
			if visited[succ] {
				continue
			}
			if pre != nil {
				pre(succ)
			}
			visit(succ)
			if post != nil {
				post(succ)
			}
		}
	}
	visit(g.Entry)
}
