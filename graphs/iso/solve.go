package iso

import (
	"fmt"
	"sort"

	"github.com/mewkiz/pkg/errutil"
)

// setPair marks the given node pair as known by removing it from c and storing
// it in m. As the graph node name is no longer a valid candidate it is removed
// from all other node pairs in c.
func (eq *equation) setPair(sname, gname string) error {
	// Sanity check.
	if key, ok := findKey(eq.m, gname); ok {
		return errutil.Newf("invalid mapping; sub node %q and %q both map to graph node %q", key, sname, gname)
	}

	// Move unique node pair from c to m.
	eq.m[sname] = gname
	delete(eq.c, sname)

	// Remove graph node name of the unique node pair from all other node
	// pairs in c.
	for key, candidates := range eq.c {
		delete(candidates, gname)
		if len(eq.c[key]) == 0 {
			return errutil.Newf("invalid mapping; sub node %q has no candidates", key)
		}
	}

	return nil
}

// findKey returns the first key in m which maps to the value val. The boolean
// value is true if such a key could be located, and false otherwise.
func findKey(m map[string]string, val string) (key string, ok bool) {
	for key, x := range m {
		if x == val {
			return key, true
		}
	}
	return "", false
}

// dup returns a copy of eq.
func (eq *equation) dup() *equation {
	// Duplicate node pair candidates.
	c := make(map[string]map[string]bool)
	for sname, candidates := range eq.c {
		c[sname] = make(map[string]bool)
		for gname, val := range candidates {
			c[sname][gname] = val
		}
	}

	// Duplicate node pairs.
	m := make(map[string]string)
	for sname, gname := range eq.m {
		m[sname] = gname
	}

	return &equation{c: c, m: m}
}

// solveUnique tries to locate a unique node pair in c. If successful the node
// pair is removed from c and stored in m. As the graph node name of the node
// pair is no longer a valid candidate it is removed from all other node pairs
// in c.
func (eq *equation) solveUnique() (ok bool, err error) {
	// Sort keys to make the algorithm deterministic.
	var snames []string
	for sname := range eq.c {
		snames = append(snames, sname)
	}
	sort.Strings(snames)

	for _, sname := range snames {
		candidates := eq.c[sname]
		if len(candidates) == 1 {
			gname := firstKey(candidates)
			err := eq.setPair(sname, gname)
			if err != nil {
				return false, errutil.Err(err)
			}
			return true, nil
		}
	}

	return false, nil
}

// firstKey returns the only key in m.
func firstKey(m map[string]bool) string {
	if len(m) != 1 {
		panic(fmt.Sprintf("invalid map length; expected 1, got %d", len(m)))
	}
	for key := range m {
		return key
	}
	panic("unreachable")
}
