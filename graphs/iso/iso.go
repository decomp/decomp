// Package iso implements subgraph isomorphism search algorithms.
package iso

import (
	"sort"

	"decomp.org/decomp/graphs"
	"github.com/mewspring/dot"
)

// Isomorphism returns a mapping from sub node name to graph node name if there
// exists an isomorphism of sub in graph which starts at the entry node. The
// boolean value is true if such a mapping could be located, and false
// otherwise.
func Isomorphism(graph *dot.Graph, entry string, sub *graphs.SubGraph) (m map[string]string, ok bool) {
	eq, err := candidates(graph, entry, sub)
	if err != nil {
		return nil, false
	}
	m, err = eq.solveBrute(graph, sub)
	if err != nil {
		return nil, false
	}
	return m, true
}

// Search tries to locate an isomorphism of sub in graph. If successful it
// returns the mapping from sub node name to graph node name of the first
// isomorphism located. The boolean value is true if such a mapping could be
// located, and false otherwise.
func Search(graph *dot.Graph, sub *graphs.SubGraph) (m map[string]string, ok bool) {
	var names []string
	for name := range graph.Nodes.Lookup {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		m, ok = Isomorphism(graph, name, sub)
		if ok {
			return m, true
		}
	}
	return nil, false
}
