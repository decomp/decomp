package hammock

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"gonum.org/v1/gonum/graph"
)

// PreLoop represents a pre-test loop.
//
// Pseudo-code:
//
//    while (A) {
//       B
//    }
//    C
type PreLoop struct {
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
func (prim PreLoop) Prim() *primitive.Primitive {
	cond, body, exit := prim.Cond.DOTID(), prim.Body.DOTID(), prim.Exit.DOTID()
	return &primitive.Primitive{
		Prim: "pre_loop",
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
//    digraph pre_loop {
//       cond -> body
//       cond -> exit
//       body -> cond
//    }
func (prim PreLoop) String() string {
	cond, body, exit := prim.Cond.DOTID(), prim.Body.DOTID(), prim.Exit.DOTID()
	const format = `
digraph pre_loop {
	%[1]s -> %[2]s
	%[1]s -> %[3]s
	%[2]s -> %[1]s
}`
	return fmt.Sprintf(format[1:], cond, body, exit)
}

// FindPreLoop returns the first occurrence of a pre-test loop in g, and a
// boolean indicating if such a primitive was found.
func FindPreLoop(g graph.Directed, dom cfa.DominatorTree) (prim PreLoop, ok bool) {
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
	return PreLoop{}, false
}

// IsValid reports whether the cond, body and exit node candidates of prim form
// a valid pre-test loop in g.
//
// Control flow graph:
//
//    cond
//    ↓  ↖↘
//    ↓   body
//    ↓
//    exit
func (prim PreLoop) IsValid(g graph.Directed, dom cfa.DominatorTree) bool {
	// Dominator sanity check.
	cond, body, exit := prim.Cond, prim.Body, prim.Exit
	if !dom.Dominates(cond.ID(), body.ID()) || !dom.Dominates(cond.ID(), exit.ID()) {
		return false
	}
	// Verify that cond has two successors (body and exit).
	condSuccs := g.From(cond.ID())
	if condSuccs.Len() != 2 || !g.HasEdgeFromTo(cond.ID(), body.ID()) || !g.HasEdgeFromTo(cond.ID(), exit.ID()) {
		return false
	}
	// Verify that body has one predecessor (cond) and one successor (cond).
	bodyPreds := g.To(body.ID())
	bodySuccs := g.From(body.ID())
	if bodyPreds.Len() != 1 || bodySuccs.Len() != 1 || !g.HasEdgeFromTo(body.ID(), cond.ID()) {
		return false
	}
	// Verify that exit has one predecessor (cond).
	exitPreds := g.To(exit.ID())
	return exitPreds.Len() == 1
}
