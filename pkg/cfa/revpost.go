package cfa

import "gonum.org/v1/gonum/graph"

// RevPostOrder returns the nodes of the graph in reverse post-order; as
// computed by performing a depth-first traversal of the control flow graph --
// starting at the entry node -- and storing nodes in post-order, and finally
// reversing the list of stored nodes.
//
// The benefit with reverse post-order is that it guarantees that each node of
// the list is present before any of its successors (not taking cycles into
// account).
func RevPostOrder(g Graph) []graph.Node {
	var ns []graph.Node
	post := func(n graph.Node) {
		ns = append(ns, n)
	}
	DFS(g, nil, post)
	reverse(ns)
	return ns
}

// reverse reverses the list of nodes in-place.
func reverse(nodes []graph.Node) []graph.Node {
	n := len(nodes)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return nodes
}
