package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"gonum.org/v1/gonum/graph"
)

// PostLoop represents a post-test loop.
//
// Pseudo-code:
//
//    do {
//    } while (A)
//    B
type PostLoop struct {
	// Condition node (A).
	Cond graph.Node
	// Exit node (B).
	Exit graph.Node
}

// Prim returns a representation of the high-level control flow primitive, as a
// mapping from control flow primitive node names to control flow graph node
// names.
//
// Example mapping:
//
//    "cond": "A"
//    "exit": "B"
func (prim PostLoop) Prim() *primitive.Primitive {
	cond, exit := label(prim.Cond), label(prim.Exit)
	return &primitive.Primitive{
		Prim: "post_loop",
		Nodes: map[string]string{
			"cond": cond,
			"exit": exit,
		},
		Entry: cond,
		Exit:  exit,
	}
}

// String returns a string representation of prim in DOT format.
//
// Example output:
//
//    digraph post_loop {
//       cond -> cond
//       cond -> exit
//    }
func (prim PostLoop) String() string {
	cond, exit := label(prim.Cond), label(prim.Exit)
	const format = `
digraph post_loop {
	%v -> %v
	%v -> %v
}`
	return fmt.Sprintf(format[1:], cond, cond, cond, exit)
}

// FindPostLoop returns the first occurrence of a post-test loop in g, and a
// boolean indicating if such a primitive was found.
func FindPostLoop(g graph.Directed, dom cfg.DominatorTree) (prim PostLoop, ok bool) {
	// Range through cond node candidates.
	condNodes := g.Nodes()
	for condNodes.Next() {
		cond := condNodes.Node()
		// Verify that cond has two successors (cond and exit).
		condSuccs := graph.NodesOf(g.From(cond.ID()))
		if len(condSuccs) != 2 {
			continue
		}
		prim.Cond = cond

		// Try the first exit node candidate.
		prim.Exit = condSuccs[0]
		if prim.IsValid(g, dom) {
			return prim, true
		}

		// Try the second exit node candidate.
		prim.Exit = condSuccs[1]
		if prim.IsValid(g, dom) {
			return prim, true
		}
	}
	return PostLoop{}, false
}

// IsValid reports whether the cond and exit node candidates of prim form a
// valid post-test loop in g.
//
// Control flow graph:
//
//    cond ↘
//    ↓   ↖↲
//    ↓
//    exit
func (prim PostLoop) IsValid(g graph.Directed, dom cfg.DominatorTree) bool {
	// Dominator sanity check.
	cond, exit := prim.Cond, prim.Exit
	if !dom.Dominates(cond, exit) {
		return false
	}

	// Verify that cond has two successors (cond and exit).
	condSuccs := g.From(cond.ID())
	if condSuccs.Len() != 2 || !g.HasEdgeFromTo(cond.ID(), cond.ID()) || !g.HasEdgeFromTo(cond.ID(), exit.ID()) {
		return false
	}

	// Verify that exit has one predecessor (cond).
	exitPreds := g.To(exit.ID())
	return exitPreds.Len() == 1
}
