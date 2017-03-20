package cfg

import (
	"strconv"
	"strings"

	"github.com/gonum/graph"
	"github.com/graphism/dot"
	"github.com/graphism/dot/ast"
	"github.com/pkg/errors"
)

// ParseFile parses the given Graphviz DOT file into a control flow graph.
func ParseFile(path string) (*Graph, error) {
	file, err := dot.ParseFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(file.Graphs) != 1 {
		return nil, errors.Errorf("invalid number of graphs in DOT file %q; expected 1, got %d", path, len(file.Graphs))
	}
	src := file.Graphs[0]
	g := newGraph()
	dot.CopyDirected(g, src)
	for _, n := range g.Nodes() {
		if n, ok := n.(*Node); ok {
			if len(n.Label) < 1 {
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

// NewNode returns a new node with a unique node ID in the graph.
func (g *Graph) NewNode() graph.Node {
	return g.newNode()
}

// NewEdge returns a new edge from the source to the destination node in the
// graph, or the existing edge if already present.
func (g *Graph) NewEdge(from, to graph.Node) graph.Edge {
	return g.newEdge(from, to)
}

// UnmarshalDOTAttr decodes a single DOT attribute.
func (n *Node) UnmarshalDOTAttr(attr *ast.Attr) error {
	switch attr.Key {
	case "label":
		s := attr.Val
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			var err error
			s, err = strconv.Unquote(attr.Val)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		n.Label = s
	default:
		return errors.Errorf("support for decoding attribute with key %q not yet implemented", attr.Key)
	}
	return nil
}

// UnmarshalDOTAttr decodes a single DOT attribute.
func (e *Edge) UnmarshalDOTAttr(attr *ast.Attr) error {
	switch attr.Key {
	case "label":
		s := attr.Val
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			var err error
			s, err = strconv.Unquote(attr.Val)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		e.Label = s
	default:
		return errors.Errorf("support for decoding attribute with key %q not yet implemented", attr.Key)
	}
	return nil
}
