package iso

import (
	"sort"

	"decomp.org/decomp/graphs"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewspring/dot"
)

// solveBrute tries to solve the node pair equation through brute force. It
// recursively locates and attempts to solve the easiest node pair (i.e. the one
// with the fewest number of candidates) until the equation is solved, or until
// all potential solutions have been exhausted.
func (eq *equation) solveBrute(graph *dot.Graph, sub *graphs.SubGraph) (m map[string]string, err error) {
	if eq.isValid(graph, sub) {
		return eq.m, nil
	}
	sname, err := eq.easiest()
	if err != nil {
		return nil, err
	}

	// Sort candidates to make the algorithm deterministic.
	candidates := make([]string, 0, len(eq.c[sname]))
	for gname := range eq.c[sname] {
		candidates = append(candidates, gname)
	}
	sort.Strings(candidates)

	for _, gname := range candidates {
		dup := eq.dup()
		err = dup.setPair(sname, gname)
		if err != nil {
			continue
		}
		m, err := dup.solveBrute(graph, sub)
		if err != nil {
			continue
		}
		return m, nil
	}

	return nil, errutil.New("unable to locate node pair mapping")
}

// easiest returns the sub node name of the easiest node pair (i.e. the one with
// the fewest number of candidates) to solve.
func (eq *equation) easiest() (string, error) {
	// Sort keys to make the algorithm deterministic.
	var snames []string
	for sname := range eq.c {
		snames = append(snames, sname)
	}
	sort.Strings(snames)

	min := -1
	var easiest string
	for _, sname := range snames {
		candidates := eq.c[sname]
		if min == -1 || len(candidates) < min {
			min = len(candidates)
			easiest = sname
		}
	}
	if min < 1 {
		return "", errutil.Newf("too few candidates for brute force; expected > 1, got %d", min)
	}
	return easiest, nil
}
