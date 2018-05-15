package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"gonum.org/v1/gonum/graph"
)

// If represents a 1-way conditional statement.
//
// Pseudo-code:
//
//    if (A) {
//       B
//    }
//    C
type If struct {
	// Condition node (A).
	Cond graph.Node
	// Body node (B).
	Body graph.Node
	// Exit node (C).
	Exit graph.Node
}

// Prim returns a representation of the high-level control flow primitive, as a
// mapping from control flow primitive node names to control flow graph node
// names.
//
// Example mapping:
//
//    "cond": "A"
//    "body": "B"
//    "exit": "C"
func (prim If) Prim() *primitive.Primitive {
	cond, body, exit := label(prim.Cond), label(prim.Body), label(prim.Exit)
	return &primitive.Primitive{
		Prim: "if",
		Nodes: map[string]string{
			"cond": cond,
			"body": body,
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
//    digraph if {
//       cond -> body
//       cond -> exit
//       body -> exit
//    }
func (prim If) String() string {
	cond, body, exit := label(prim.Cond), label(prim.Body), label(prim.Exit)
	const format = `
digraph if {
	%v -> %v
	%v -> %v
	%v -> %v
}`
	return fmt.Sprintf(format[1:], cond, body, cond, exit, body, exit)
}

// FindIf returns the first occurrence of a 1-way conditional statement in g,
// and a boolean indicating if such a primitive was found.
func FindIf(g graph.Directed, dom cfg.DominatorTree) (prim If, ok bool) {
	// Range through cond node candidates.
	for _, cond := range g.Nodes() {
		// Verify that cond has two successors (body and exit).
		condSuccs := g.From(cond.ID())
		if len(condSuccs) != 2 {
			continue
		}
		prim.Cond = cond

		// Select body and exit node candidates.
		prim.Body, prim.Exit = condSuccs[0], condSuccs[1]
		if prim.IsValid(g, dom) {
			return prim, true
		}

		// Swap body and exit node candidates and try again.
		prim.Body, prim.Exit = prim.Exit, prim.Body
		if prim.IsValid(g, dom) {
			return prim, true
		}
	}
	return If{}, false
}

// IsValid reports whether the cond, body and exit node candidates of prim form
// a valid 1-way conditional statement in g.
//
// Control flow graph:
//
//    cond
//    ↓   ↘
//    ↓    body
//    ↓   ↙
//    exit
func (prim If) IsValid(g graph.Directed, dom cfg.DominatorTree) bool {
	// Dominator sanity check.
	cond, body, exit := prim.Cond, prim.Body, prim.Exit
	if !dom.Dominates(cond, body) || !dom.Dominates(cond, exit) {
		return false
	}

	// Verify that cond has two successors (body and exit).
	condSuccs := g.From(cond.ID())
	if len(condSuccs) != 2 || !g.HasEdgeFromTo(cond.ID(), body.ID()) || !g.HasEdgeFromTo(cond.ID(), exit.ID()) {
		return false
	}

	// Verify that body has one predecessor (cond) and one successor (exit).
	bodyPreds := g.To(body.ID())
	bodySuccs := g.From(body.ID())
	if len(bodyPreds) != 1 || len(bodySuccs) != 1 || !g.HasEdgeFromTo(body.ID(), exit.ID()) {
		return false
	}

	// Verify that exit has two predecessors (cond and body).
	exitPreds := g.To(exit.ID())
	return len(exitPreds) == 2
}
