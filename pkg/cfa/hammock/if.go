package hammock

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
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
	Cond cfa.Node
	// Body node (B).
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
func (prim If) Prim() *primitive.Primitive {
	cond, body, exit := prim.Cond.DOTID(), prim.Body.DOTID(), prim.Exit.DOTID()
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

// String returns a string representation of prim in Graphviz DOT format.
//
// Example output:
//
//    digraph if {
//       cond -> body
//       cond -> exit
//       body -> exit
//    }
func (prim If) String() string {
	cond, body, exit := prim.Cond.DOTID(), prim.Body.DOTID(), prim.Exit.DOTID()
	const format = `
digraph if {
	%[1]s -> %[2]s
	%[1]s -> %[3]s
	%[2]s -> %[3]s
}`
	return fmt.Sprintf(format[1:], cond, body, exit)
}

// FindIf returns the first occurrence of a 1-way conditional statement in g,
// and a boolean indicating if such a primitive was found.
func FindIf(g graph.Directed, dom cfa.DominatorTree) (prim If, ok bool) {
	// Range through cond node candidates.
	for nodes := g.Nodes(); nodes.Next(); {
		// Note: This run-time type assertion goes away, should Gonum graph start
		// to leverage generics in Go2.
		cond := nodes.Node().(cfa.Node)
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
func (prim If) IsValid(g graph.Directed, dom cfa.DominatorTree) bool {
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
	// Verify that body has one predecessor (cond) and one successor (exit).
	bodyPreds := g.To(body.ID())
	bodySuccs := g.From(body.ID())
	if bodyPreds.Len() != 1 || bodySuccs.Len() != 1 || !g.HasEdgeFromTo(body.ID(), exit.ID()) {
		return false
	}
	// Verify that exit has two predecessors (cond and body).
	exitPreds := g.To(exit.ID())
	return exitPreds.Len() == 2
}
