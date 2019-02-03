package hammock

import (
	"fmt"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
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
	Cond cfa.Node
	// Exit node (B).
	Exit cfa.Node
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
	cond, exit := prim.Cond.DOTID(), prim.Exit.DOTID()
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
	cond, exit := prim.Cond.DOTID(), prim.Exit.DOTID()
	const format = `
digraph post_loop {
	%[1]s -> %[1]s
	%[1]s -> %[2]s
}`
	return fmt.Sprintf(format[1:], cond, exit)
}

// FindPostLoop returns the first occurrence of a post-test loop in g, and a
// boolean indicating if such a primitive was found.
func FindPostLoop(g graph.Directed, dom cfa.DominatorTree) (prim PostLoop, ok bool) {
	// Range through cond node candidates.
	for nodes := g.Nodes(); nodes.Next(); {
		cond := nodes.Node().(cfa.Node)
		// Verify that cond has two successors (cond and exit).
		condSuccs := cfa.NodesOf(g.From(cond.ID()))
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
func (prim PostLoop) IsValid(g graph.Directed, dom cfa.DominatorTree) bool {
	// Dominator sanity check.
	cond, exit := prim.Cond, prim.Exit
	if !dom.Dominates(cond.ID(), exit.ID()) {
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
