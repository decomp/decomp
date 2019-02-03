package hammock

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"gonum.org/v1/gonum/graph"
)

// Seq represents a sequence of two statements.
//
// Pseudo-code:
//
//    A
//    B
type Seq struct {
	// Entry node (A).
	Entry cfa.Node
	// Exit node (B).
	Exit cfa.Node
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
	entry, exit := prim.Entry.DOTID(), prim.Exit.DOTID()
	return &primitive.Primitive{
		Prim: "seq",
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
	entry, exit := prim.Entry.DOTID(), prim.Exit.DOTID()
	const format = `
digraph seq {
	%s -> %s
}`
	return fmt.Sprintf(format[1:], entry, exit)
}

// FindSeq returns the first occurrence of a sequence of two statements in g,
// and a boolean indicating if such a primitive was found.
func FindSeq(g graph.Directed, dom cfa.DominatorTree) (prim Seq, ok bool) {
	// Range through entry node candidates.
	for nodes := g.Nodes(); nodes.Next(); {
		// Note: This run-time type assertion goes away, should Gonum graph start
		// to leverage generics in Go2.
		entry := nodes.Node().(cfa.Node)
		// Verify that entry has one successor (exit).
		entrySuccs := cfa.NodesOf(g.From(entry.ID()))
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
func (prim Seq) IsValid(g graph.Directed, dom cfa.DominatorTree) bool {
	// Dominator sanity check.
	entry, exit := prim.Entry, prim.Exit
	if !dom.Dominates(entry.ID(), exit.ID()) {
		return false
	}

	// Verify that entry has one successor (exit).
	entrySuccs := g.From(entry.ID())
	if entrySuccs.Len() != 1 || !g.HasEdgeFromTo(entry.ID(), exit.ID()) {
		return false
	}

	// Verify that exit has one predecessor (entry).
	exitPreds := g.To(exit.ID())
	return exitPreds.Len() == 1
}
