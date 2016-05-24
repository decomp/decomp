package iso

import (
	"fmt"
	"sort"

	"decomp.org/x/graphs"
	"github.com/mewfork/dot"
)

// isValid returns true if m is a valid mapping, from sub node name to graph
// node name, for an isomorphism of sub in graph considering all nodes and edges
// except predecessors of entry and successors of exit.
func (eq *equation) isValid(graph *dot.Graph, sub *graphs.SubGraph) bool {
	if len(eq.m) != len(sub.Nodes.Nodes) {
		return false
	}

	// Check for duplicate values.
	if hasDup(eq.m) {
		return false
	}

	// Verify that the entry node dominates the exit node.
	entry, ok := graph.Nodes.Lookup[eq.m[sub.Entry()]]
	if !ok {
		return false
	}
	exit, ok := graph.Nodes.Lookup[eq.m[sub.Exit()]]
	if !ok {
		return false
	}
	// TODO: Figure out how to handle find the if-statement in the following graph:
	//    digraph bar {
	//       E -> F
	//       E -> J
	//       F -> G
	//       F -> E
	//       G -> I
	//       I -> E
	//       E [label="entry"]
	//       F
	//       G
	//       I
	//       J [label="exit"]
	//    }
	//
	// ref: https://github.com/decomp/decompilation/issues/172
	if !entry.Dominates(exit) {
		return false
	}

	// Sort keys to make the algorithm deterministic.
	var snames []string
	for sname := range eq.m {
		snames = append(snames, sname)
	}
	sort.Strings(snames)

	for _, sname := range snames {
		gname := eq.m[sname]
		s, ok := sub.Nodes.Lookup[sname]
		if !ok {
			panic(fmt.Sprintf("unable to locate node %q in sub", sname))
		}
		g, ok := graph.Nodes.Lookup[gname]
		if !ok {
			panic(fmt.Sprintf("unable to locate node %q in graph", gname))
		}

		// Verify predecessors.
		if s.Name != sub.Entry() {
			if len(s.Preds) != len(g.Preds) {
				return false
			}
			for _, spred := range s.Preds {
				found := false
				for _, gpred := range g.Preds {
					if gpred.Name == eq.m[spred.Name] {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}

		// Verify successors.
		if s.Name != sub.Exit() {
			if len(s.Succs) != len(g.Succs) {
				return false
			}
			for _, ssucc := range s.Succs {
				found := false
				for _, gsucc := range g.Succs {
					if gsucc.Name == eq.m[ssucc.Name] {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	// Isomorphism found!
	return true
}

// hasDup returns true if m contains a duplicate value.
func hasDup(m map[string]string) bool {
	vals := make(map[string]bool, len(m))
	for _, v := range m {
		if vals[v] {
			return true
		}
		vals[v] = true
	}
	return false
}
