package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"github.com/gonum/graph"
)

// Seq represents a sequence of two statements.
//
// Pseudo-code:
//
//    A
//    B
type Seq struct {
	// Entry node (A).
	Entry graph.Node
	// Exit node (B).
	Exit graph.Node
}

// Prim returns a representation of the high-level control flow primitive, as a
// mapping from control flow primitive node names to control flow graph node
// names.
//
// Example mapping:
//
//    "entry": "A"
//    "exit":  "B"
func (prim Seq) Prim() *primitive.Primitive {
	entry, exit := label(prim.Entry), label(prim.Exit)
	return &primitive.Primitive{
		Prim: "seq",
		// Note, the primitive node name should be set to a unique node ID when
		// merged into the CFG.
		Node: "",
		Nodes: map[string]string{
			"entry": entry,
			"exit":  exit,
		},
		Entry: entry,
		Exit:  exit,
	}
}

// String returns a string representation of prim in DOT format.
//
// Example output:
//
//    digraph seq {
//       entry -> exit
//    }
func (prim Seq) String() string {
	entry, exit := label(prim.Entry), label(prim.Exit)
	const format = `
digraph seq {
	%v -> %v
}`
	return fmt.Sprintf(format[1:], entry, exit)
}

// FindSeq returns the first occurrence of a sequence of two statements in g,
// and a boolean indicating if such a primitive was found.
func FindSeq(g graph.Directed, dom cfg.Dom) (prim Seq, ok bool) {
	// Range through entry node candidates.
	for _, entry := range g.Nodes() {
		// Verify that entry has one successor (exit).
		entrySuccs := g.From(entry)
		if len(entrySuccs) != 1 {
			continue
		}
		prim.Entry = entry

		// Select exit node candidate.
		prim.Exit = entrySuccs[0]
		if prim.IsValid(g, dom) {
			return prim, true
		}
	}
	return Seq{}, false
}

// IsValid reports whether the entry and exit node candidates of prim form a
// valid sequence of two statements in g.
//
// Control flow graph:
//
//    entry
//    ↓
//    exit
func (prim Seq) IsValid(g graph.Directed, dom cfg.Dom) bool {
	// Dominator sanity check.
	entry, exit := prim.Entry, prim.Exit
	if !dom.Dominates(entry, exit) {
		return false
	}

	// Verify that entry has one successor (exit).
	entrySuccs := g.From(entry)
	if len(entrySuccs) != 1 || !g.HasEdgeFromTo(entry, exit) {
		return false
	}

	// Verify that exit has one predecessor (entry).
	exitPreds := g.To(exit)
	return len(exitPreds) == 1
}
