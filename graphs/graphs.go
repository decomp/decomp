// Package graphs provides access to subgraph data structures.
package graphs

import (
	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
)

// SubGraph represents a subgraph with a dedicated entry and exit node. Incoming
// edges to entry and outgoing edges from exit are ignored when searching for
// isomorphisms of the subgraph.
type SubGraph struct {
	*dot.Graph
	entry, exit string
}

// ParseSubGraph parses the provided DOT file into a subgraph with a dedicated
// entry and exit node. The entry and exit nodes are identified using the node
// "label" attribute, e.g.
//
//    digraph if {
//       A->B [label="true"]
//       A->C [label="false"]
//       B->C
//       A [label="entry"]
//       B
//       C [label="exit"]
//    }
func ParseSubGraph(path string) (*SubGraph, error) {
	graph, err := dot.ParseFile(path)
	if err != nil {
		return nil, err
	}
	return NewSubGraph(graph)
}

// NewSubGraph returns a new subgraph based on graph with a dedicated entry and
// exit node. The entry and exit nodes are identified using the node "label"
// attribute, e.g.
//
//    digraph if {
//       A->B [label="true"]
//       A->C [label="false"]
//       B->C
//       A [label="entry"]
//       B
//       C [label="exit"]
//    }
func NewSubGraph(graph *dot.Graph) (*SubGraph, error) {
	sub := &SubGraph{Graph: graph}

	// Locate entry and exit nodes.
	var hasEntry, hasExit bool
	for _, node := range graph.Nodes.Nodes {
		label, ok := node.Attrs["label"]
		if !ok {
			continue
		}
		switch label {
		case "entry":
			if hasEntry {
				return nil, errutil.Newf(`redefinition of node with "entry" label; previous node %q, new node %q`, sub.entry, node.Name)
			}
			sub.entry = node.Name
			hasEntry = true
		case "exit":
			if hasExit {
				return nil, errutil.Newf(`redefinition of node with "exit" label; previous node %q, new node %q`, sub.exit, node.Name)
			}
			sub.exit = node.Name
			hasExit = true
		}
	}
	if !hasEntry {
		return nil, errutil.New(`unable to locate node with "entry" label`)
	}
	if !hasExit {
		return nil, errutil.New(`unable to locate node with "exit" label`)
	}

	return sub, nil
}

// Entry returns the entry node name in the subgraph.
func (sub *SubGraph) Entry() string {
	return sub.entry
}

// Exit returns the exit node name in the subgraph.
func (sub *SubGraph) Exit() string {
	return sub.exit
}
