package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"gonum.org/v1/gonum/graph"
)

// IfReturn represents a 1-way conditional with a body return statement.
//
// Pseudo-code:
//
//    if (A) {
//       B
//       return
//    }
//    C
type IfReturn struct {
	// Condition node (A).
	Cond graph.Node
	// Body node with return statement (B).
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
func (prim IfReturn) Prim() *primitive.Primitive {
	cond, body, exit := label(prim.Cond), label(prim.Body), label(prim.Exit)
	return &primitive.Primitive{
		Prim: "if_return",
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
//    digraph if_return {
//       cond -> body
//       cond -> exit
//    }
func (prim IfReturn) String() string {
	cond, body, exit := label(prim.Cond), label(prim.Body), label(prim.Exit)
	const format = `
digraph if_return {
	%v -> %v
	%v -> %v
}`
	return fmt.Sprintf(format[1:], cond, body, cond, exit)
}

// FindIfReturn returns the first occurrence of a 1-way conditional with a body
// return statement in g, and a boolean indicating if such a primitive was
// found.
func FindIfReturn(g graph.Directed, dom cfg.DominatorTree) (prim IfReturn, ok bool) {
	// Range through cond node candidates.
	for _, cond := range g.Nodes() {
		// Verify that cond has two successors (body and exit).
		condSuccs := g.From(cond)
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
	return IfReturn{}, false
}

// IsValid reports whether the cond, body and exit node candidates of prim form
// a valid 1-way conditional with a body return statement in g.
//
// Control flow graph:
//
//    cond
//    ↓   ↘
//    ↓    body
//    ↓
//    exit
func (prim IfReturn) IsValid(g graph.Directed, dom cfg.DominatorTree) bool {
	// Dominator sanity check.
	cond, body, exit := prim.Cond, prim.Body, prim.Exit
	if !dom.Dominates(cond, body) || !dom.Dominates(cond, exit) {
		return false
	}

	// Verify that cond has two successors (body and exit).
	condSuccs := g.From(cond)
	if len(condSuccs) != 2 || !g.HasEdgeFromTo(cond, body) || !g.HasEdgeFromTo(cond, exit) {
		return false
	}

	// Verify that body has one predecessor (cond) and zero successors.
	bodyPreds := g.To(body)
	bodySuccs := g.From(body)
	if len(bodyPreds) != 1 || len(bodySuccs) != 0 {
		return false
	}

	// Verify that exit has one predecessor (cond).
	exitPreds := g.To(exit)
	if len(exitPreds) != 1 {
		return false
	}

	// Verify that the entry node (cond) has no predecessors dominated by cond,
	// as that would indicate a loop construct.
	//
	//       cond
	//     ↗ ↓   ↘
	//    ↑  ↓    body
	//    ↑  ↓
	//    ↑  exit
	//     ↖ ↓
	//       A
	condPreds := g.To(cond)
	for _, pred := range condPreds {
		if dom.Dominates(cond, pred) {
			return false
		}
	}

	return true
}
