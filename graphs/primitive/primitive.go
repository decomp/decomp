// Package primitive defines the types used to represent high-level control flow
// primitives.
package primitive

// A Primitive represents a high-level control flow primitive (e.g. 2-way
// conditional, pre-test loop) as a mapping from subgraph (graph representation
// of a control flow primitive) node names to control flow graph node names.
type Primitive struct {
	// Primitive name; e.g. "if", "pre_loop", ...
	Prim string `json:"prim"`
	// Node name of the primitive; e.g. "if_17".
	Node string `json:"node"`
	// Node mapping; e.g. {"cond": "17", "body": "24", "exit": "32"}
	Nodes map[string]string `json:"nodes"`
	// Entry node.
	Entry string `json:"entry"`
	// Exit node.
	Exit string `json:"exit,omitempty"`
}
