package hammock

import (
	"fmt"

	"github.com/mewmew/lnp/cfa"
	"github.com/mewmew/lnp/cfa/primitive"
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
	Cond cfa.Node
	// Body node with return statement (B).
	Body cfa.Node
	// Exit node (C).
	Exit cfa.Node
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
	cond, body, exit := prim.Cond.DOTID(), prim.Body.DOTID(), prim.Exit.DOTID()
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
	cond, body, exit := prim.Cond.DOTID(), prim.Body.DOTID(), prim.Exit.DOTID()
	const format = `
digraph if_return {
	%[1]s -> %[2]s
	%[1]s -> %[3]s
}`
	return fmt.Sprintf(format[1:], cond, body, exit)
}

// FindIfReturn returns the first occurrence of a 1-way conditional with a body
// return statement in g, and a boolean indicating if such a primitive was
// found.
func FindIfReturn(g graph.Directed, dom cfa.DominatorTree) (prim IfReturn, ok bool) {
	// Range through cond node candidates.
	condNodes := g.Nodes()
	for condNodes.Next() {
		// Note: This run-time type assertion goes away, should Gonum graph start
		// to leverage generics in Go2.
		cond := condNodes.Node().(cfa.Node)
		// Verify that cond has two successors (body and exit).
		condSuccs := cfa.NodesOf(g.From(cond.ID()))
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
		prim.Body, prim.Exit = condSuccs[1], condSuccs[0]
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
func (prim IfReturn) IsValid(g graph.Directed, dom cfa.DominatorTree) bool {
	// Dominator sanity check.
	cond, body, exit := prim.Cond, prim.Body, prim.Exit
	if !dom.Dominates(cond, body) || !dom.Dominates(cond, exit) {
		return false
	}
	// Verify that cond has two successors (body and exit).
	condSuccs := g.From(cond.ID())
	if condSuccs.Len() != 2 || !g.HasEdgeFromTo(cond.ID(), body.ID()) || !g.HasEdgeFromTo(cond.ID(), exit.ID()) {
		return false
	}
	// Verify that body has one predecessor (cond) and zero successors.
	bodyPreds := g.To(body.ID())
	bodySuccs := g.From(body.ID())
	if bodyPreds.Len() != 1 || bodySuccs.Len() != 0 {
		return false
	}
	// Verify that exit has one predecessor (cond).
	exitPreds := g.To(exit.ID())
	if exitPreds.Len() != 1 {
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
	condPreds := g.To(cond.ID())
	for condPreds.Next() {
		pred := condPreds.Node()
		if dom.Dominates(cond, pred) {
			return false
		}
	}
	return true
}
