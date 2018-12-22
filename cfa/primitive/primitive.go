// Package primitive defines the types used to represent high-level control flow
// primitives.
package primitive

import (
	"fmt"
	"strings"

	"github.com/rickypai/natsort"
)

// A Primitive represents a high-level control flow primitive (e.g. 2-way
// conditional, pre-test loop) as a mapping from subgraph node names to control
// flow graph node names, where the subgraph is the cannonical representation of
// a given high-level control flow primitive.
type Primitive struct {
	// Primitive name; e.g.
	//
	//    "if", "pre_loop", ...
	Prim string `json:"prim"`
	// Node mapping; e.g.
	//
	//    {"cond": "17", "body": "24", "exit": "32"}
	Nodes map[string]string `json:"nodes"`
	// Entry node name.
	Entry string `json:"entry"`
	// Exit node name.
	Exit string `json:"exit,omitempty"`
}

// String returns the string representation of the high-level control flow
// primitive.
func (p *Primitive) String() string {
	buf := &strings.Builder{}
	var keys []string
	for key := range p.Nodes {
		keys = append(keys, key)
	}
	natsort.Strings(keys)
	fmt.Fprintf(buf, "prim: %s\n", p.Prim)
	buf.WriteString("nodes:\n")
	for _, key := range keys {
		fmt.Fprintf(buf, "   %s: %s\n", key, p.Nodes[key])
	}
	fmt.Fprintf(buf, "entry: %s\n", p.Entry)
	if len(p.Exit) > 0 {
		fmt.Fprintf(buf, "exit: %s", p.Exit)
	}
	return buf.String()
}
