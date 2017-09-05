package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"gonum.org/v1/gonum/graph"
)

// IfElse represents a 2-way conditional statement.
//
// Pseudo-code:
//
//    if (A) {
//       B
//    } else {
//       C
//    }
//    D
type IfElse struct {
	// Condition node (A).
	Cond graph.Node
	// Body node of the true branch (B).
	BodyTrue graph.Node
	// Body node of the false branch (C).
	BodyFalse graph.Node
	// Exit node (D).
	Exit graph.Node
}

// Prim returns a representation of the high-level control flow primitive, as a
// mapping from control flow primitive node names to control flow graph node
// names.
//
// Example mapping:
//
//    "cond":       "A"
//    "body_true":  "B"
//    "body_false": "C"
//    "exit":       "D"
func (prim IfElse) Prim() *primitive.Primitive {
	cond, bodyTrue, bodyFalse, exit := label(prim.Cond), label(prim.BodyTrue), label(prim.BodyFalse), label(prim.Exit)
	return &primitive.Primitive{
		Prim: "if_else",
		Nodes: map[string]string{
			"cond":       cond,
			"body_true":  bodyTrue,
			"body_false": bodyFalse,
			"exit":       exit,
		},
		Entry: cond,
		Exit:  exit,
	}
}

// String returns a string representation of prim in DOT format.
//
// Example output:
//
//    digraph if_else {
//       cond -> body_true
//       cond -> body_false
//       body_true -> exit
//       body_false -> exit
//    }
func (prim IfElse) String() string {
	cond, bodyTrue, bodyFalse, exit := label(prim.Cond), label(prim.BodyTrue), label(prim.BodyFalse), label(prim.Exit)
	const format = `
digraph if_else {
	%v -> %v
	%v -> %v
	%v -> %v
	%v -> %v
}`
	return fmt.Sprintf(format[1:], cond, bodyTrue, cond, bodyFalse, bodyTrue, exit, bodyFalse, exit)
}

// FindIfElse returns the first occurrence of a 2-way conditional statement in
// g, and a boolean indicating if such a primitive was found.
func FindIfElse(g graph.Directed, dom cfg.DominatorTree) (prim IfElse, ok bool) {
	// Range through cond node candidates.
	for _, cond := range g.Nodes() {
		// Verify that cond has two successors (body_true and body_false).
		condSuccs := g.From(cond)
		if len(condSuccs) != 2 {
			continue
		}
		prim.Cond = cond

		// Select body_true and body_false node candidates.
		prim.BodyTrue, prim.BodyFalse = condSuccs[0], condSuccs[1]

		// Verify that body_true has one successor (exit).
		bodyTrueSuccs := g.From(prim.BodyTrue)
		if len(bodyTrueSuccs) != 1 {
			continue
		}

		// Select exit node candidate.
		prim.Exit = bodyTrueSuccs[0]
		if prim.IsValid(g, dom) {
			return prim, true
		}
	}
	return IfElse{}, false
}

// IsValid reports whether the cond, body_true, body_false and exit node
// candidates of prim form a valid 2-way conditional statement in g.
//
// Control flow graph:
//
//              cond
//             ↙    ↘
//    body_true      body_false
//             ↘    ↙
//              exit
func (prim IfElse) IsValid(g graph.Directed, dom cfg.DominatorTree) bool {
	// Dominator sanity check.
	cond, bodyTrue, bodyFalse, exit := prim.Cond, prim.BodyTrue, prim.BodyFalse, prim.Exit
	if !dom.Dominates(cond, bodyTrue) || !dom.Dominates(cond, bodyFalse) || !dom.Dominates(cond, exit) {
		return false
	}

	// Verify that cond has two successors (body_true and body_false).
	condSuccs := g.From(cond)
	if len(condSuccs) != 2 || !g.HasEdgeFromTo(cond, bodyTrue) || !g.HasEdgeFromTo(cond, bodyFalse) {
		return false
	}

	// Verify that body_true has one predecessor (cond) and one successor (exit).
	bodyTrueSuccs := g.From(bodyTrue)
	bodyTruePreds := g.To(bodyTrue)
	if len(bodyTruePreds) != 1 || len(bodyTrueSuccs) != 1 || !g.HasEdgeFromTo(bodyTrue, exit) {
		return false
	}

	// Verify that body_false has one predecessor (cond) and one successor (exit).
	bodyFalseSuccs := g.From(bodyFalse)
	bodyFalsePreds := g.To(bodyFalse)
	if len(bodyFalsePreds) != 1 || len(bodyFalseSuccs) != 1 || !g.HasEdgeFromTo(bodyFalse, exit) {
		return false
	}

	// Verify that exit has two predecessor (body_true and body_false).
	exitPreds := g.To(exit)
	return len(exitPreds) == 2
}
