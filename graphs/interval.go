// Intervals
//
// Definition 57: Given a node h, an interval I(h) is the maximal, single-entry
// subgraph in which h is the only entry node and in which all closed paths
// contain h. The unique interval node h is called the interval head or simply
// the header node.
//
// ref: "Reverse Compilation Techniques" by C. Cifuentes (p. 132).

package graphs

import (
	"log"

	"github.com/mewspring/dot"
)

// An Interval is a single-entry subgraph in which h is the only entry node, and
// in which all closed paths contain h. Each node in the interval is dominated
// by the entry node.
type Interval struct {
	*dot.Graph
	head *dot.Node
}

// Head returns the header node of the interval.
func (i *Interval) Head() *dot.Node {
	return i.head
}

// GetIntervals returns each interval in g with two or more nodes.
func GetIntervals(g *dot.Graph) (is []*Interval) {
	for _, h := range g.Nodes.Nodes {
		if i, ok := GetInterval(g, h); ok {
			is = append(is, i)
		}
	}
	return is
}

// GetInterval returns the interval in g with the entry node h. The boolean
// value is false if no interval with two or more nodes could be located.
func GetInterval(g *dot.Graph, h *dot.Node) (i *Interval, ok bool) {
	// Create a new interval with the entry node h.
	i = &Interval{
		Graph: dot.NewGraph(),
		head:  h,
	}
	i.Directed = true
	i.Name = g.Name + "_interval_" + h.Name
	// TODO: Add dot.Attrs{"label": "entry"}?
	i.AddNode(i.Name, h.Name, nil)

	// Add each node dominated by h to the interval.
	for _, n := range g.Nodes.Nodes {
		if h == n {
			continue
		}
		if h.Dominates(n) {
			i.AddNode(i.Name, n.Name, nil)
			ok = true
		}
	}
	if !ok {
		return nil, false
	}

	// Add the original edges between the nodes of the interval.
	for _, m := range i.Nodes.Nodes {
		n, ok := g.Nodes.Lookup[m.Name]
		if !ok {
			log.Fatalf("unable to locate interval node %q in graph", m.Name)
		}
		for _, e := range g.Edges.SrcToDsts[n.Name] {
			if _, ok := i.Nodes.Lookup[e.Src]; !ok {
				continue
			}
			if _, ok := i.Nodes.Lookup[e.Dst]; !ok {
				continue
			}
			i.AddEdge(e.Src, e.Dst, true, e.Attrs)
		}
	}

	return i, true
}
