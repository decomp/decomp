package cfg

import (
	"io/ioutil"

	"github.com/graphism/simple"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// ParseFile parses the given Graphviz DOT file into a control flow graph.
func ParseFile(path string) (*Graph, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	g := &Graph{
		DirectedGraph: simple.NewDirectedGraph(),
		nodes:         make(map[string]*Node),
	}
	if err := dot.Unmarshal(data, g); err != nil {
		return nil, errors.WithStack(err)
	}
	for _, n := range g.Nodes() {
		if n, ok := n.(*Node); ok {
			if n.entry {
				// Store entry node.
				g.entry = n
			}
			if len(n.Label) == 0 {
				return nil, errors.Errorf("invalid node %#v; missing node label", n)
			}
			if prev, ok := g.nodes[n.Label]; ok {
				return nil, errors.Errorf("more than one node with node label %q; prev %#v, new %#v", n.Label, prev, n)
			}
			g.nodes[n.Label] = n
		}
	}
	return g, nil
}
