package cfa

// DFS performs a depth-first search of the control flow graph, starting at the
// entry node. The functions pre and post are invoked if non-nil during pre- and
// post-order traversal of the graph, respectively.
func DFS(g Graph, pre, post func(n Node)) {
	visited := make(map[int64]bool)
	var visit func(n Node)
	visit = func(n Node) {
		visited[n.ID()] = true
		for succs := g.From(n.ID()); succs.Next(); {
			succ := succs.Node().(Node)
			if visited[succ.ID()] {
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
	visit(g.Entry())
}
