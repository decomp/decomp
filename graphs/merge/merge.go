// Package merge implements merging of subgraphs in graphs.
package merge

import (
	"fmt"

	"decomp.org/x/graphs"
	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
)

// Merge merges the nodes of the isomorphism of sub in graph into a single node.
// If successful it returns the name of the new node.
func Merge(graph *dot.Graph, m map[string]string, sub *graphs.SubGraph) (name string, err error) {
	var nodes []*dot.Node
	for _, gname := range m {
		node, ok := graph.Nodes.Lookup[gname]
		if !ok {
			return "", errutil.Newf("unable to locate mapping for node %q", gname)
		}
		nodes = append(nodes, node)
	}
	name = uniqName(graph, sub.Name)
	entry, ok := graph.Nodes.Lookup[m[sub.Entry()]]
	if !ok {
		return "", errutil.Newf("unable to locate mapping for entry node %q", sub.Entry())
	}
	exit, ok := graph.Nodes.Lookup[m[sub.Exit()]]
	if !ok {
		return "", errutil.Newf("unable to locate mapping for exit node %q", sub.Exit())
	}
	err = graph.Replace(nodes, name, entry, exit)
	if err != nil {
		return "", errutil.Err(err)
	}
	return name, nil
}

// uniqName returns name with a uniq numeric suffix.
func uniqName(graph *dot.Graph, name string) string {
	for id := 0; ; id++ {
		s := fmt.Sprintf("%s%d", name, id)
		_, ok := graph.Nodes.Lookup[s]
		if !ok {
			return s
		}
	}
}
