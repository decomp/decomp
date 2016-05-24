package iso

import (
	"fmt"

	"decomp.org/x/graphs"
	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
)

// Equation specifies an equation of node pair candidates and known node pairs.
type equation struct {
	// mapping from sub node name to graph node name candidates.
	c map[string]map[string]bool
	// mapping from sub node name to graph node name.
	m map[string]string
}

// candidates locates node pair candidates for an isomorphism of sub in graph
// which starts at the entry node.
func candidates(graph *dot.Graph, entry string, sub *graphs.SubGraph) (*equation, error) {
	// Sanity checks.
	g, ok := graph.Nodes.Lookup[entry]
	if !ok {
		return nil, errutil.Newf("unable to locate entry node %q in graph", entry)
	}
	s, ok := sub.Nodes.Lookup[sub.Entry()]
	if !ok {
		panic(fmt.Sprintf("unable to locate entry node %q in sub", sub.Entry()))
	}
	if !isPotential(g, s, sub) {
		return nil, errutil.Newf("invalid entry node candidate %q; expected %d successors, got %d", g.Name, len(s.Succs), len(g.Succs))
	}

	// Locate candidate node pairs.
	eq := &equation{
		c: make(map[string]map[string]bool),
		m: make(map[string]string),
	}
	eq.findCandidates(g, s, sub)
	if len(eq.c) != len(sub.Nodes.Nodes) {
		return nil, errutil.Newf("incomplete candidate mapping; expected %d map entites, got %d", len(sub.Nodes.Nodes), len(eq.c))
	}

	return eq, nil
}

// findCandidates recursively locates potential node pairs (g and s) for an
// isomorphism of sub in graph and adds them to c.
func (eq *equation) findCandidates(g, s *dot.Node, sub *graphs.SubGraph) {
	// Exit early for impossible node pairs.
	if !isPotential(g, s, sub) {
		return
	}

	// Prevent infinite cycles.
	if _, ok := eq.c[s.Name]; ok {
		if eq.c[s.Name][g.Name] {
			return
		}
	}

	// Add node pair candidate.
	if _, ok := eq.c[s.Name]; !ok {
		eq.c[s.Name] = make(map[string]bool)
	} else if s.Name == sub.Entry() {
		// Locate candidates for the entry node and its immediate successors
		// exactly once.
		return
	}
	eq.c[s.Name][g.Name] = true

	// Recursively locate candidate successor pairs.
	for _, ssucc := range s.Succs {
		for _, gsucc := range g.Succs {
			eq.findCandidates(gsucc, ssucc, sub)
		}
	}
}

// isPotential returns true if the graph node g is a potential candidate for the
// sub node s, and false otherwise.
func isPotential(g, s *dot.Node, sub *graphs.SubGraph) bool {
	// Verify predecessors.
	if s.Name != sub.Entry() && len(g.Preds) != len(s.Preds) {
		return false
	}
	// Verify successors.
	if s.Name != sub.Exit() && len(g.Succs) != len(s.Succs) {
		return false
	}
	return true
}
