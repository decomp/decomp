package cfg

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
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
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
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
	if g.entry == nil {
		panic(fmt.Errorf(`unable to locate entry node; missing DOT node label attribute "entry"`))
	}
	return g, nil
}
