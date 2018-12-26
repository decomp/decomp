// TODO: validate implementation of FindSwitch.

package hammock

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"gonum.org/v1/gonum/graph"
)

// Switch represents an n-way conditional statement.
//
// Pseudo-code:
//
//    switch (A) {
//    case X:
//       B
//    case Y:
//       C
//    case Z:
//       D
//    default:
//       E
//    }
//    F
type Switch struct {
	// Condition node (A).
	Cond cfa.Node
	// Case nodes (B, C, D, ...).
	Cases []cfa.Node
	// Default node (E).
	Default cfa.Node
	// Exit node (F).
	Exit cfa.Node
}

// Prim returns a representation of the high-level control flow primitive, as a
// mapping from control flow primitive node names to control flow graph node
// names.
//
// Example mapping:
//
//    "cond":    "A"
//    "cases":   "B,C,D"
//    "default": "E"
//    "exit":    "F"
func (prim Switch) Prim() *primitive.Primitive {
	cond, defaultLabel, exit := prim.Cond.DOTID(), prim.Default.DOTID(), prim.Exit.DOTID()
	var cases []string
	for _, c := range prim.Cases {
		cases = append(cases, c.DOTID())
	}
	return &primitive.Primitive{
		Prim: "switch",
		Nodes: map[string]string{
			"cond":    cond,
			"cases":   strings.Join(cases, ","),
			"default": defaultLabel,
			"exit":    exit,
		},
		Entry: cond,
		Exit:  exit,
	}
}

// String returns a string representation of prim in DOT format.
//
// Example output:
//
//    digraph switch {
//       cond -> case_B
//       cond -> case_C
//       cond -> case_D
//       cond -> default
//       case_B -> exit
//       case_C -> exit
//       case_D -> exit
//       default -> exit
//    }
func (prim Switch) String() string {
	cond, defaultLabel, exit := prim.Cond.DOTID(), prim.Default.DOTID(), prim.Exit.DOTID()
	var cases []string
	for _, c := range prim.Cases {
		cases = append(cases, c.DOTID())
	}
	buf := &bytes.Buffer{}
	buf.WriteString("digraph switch {\n")
	for _, c := range cases {
		fmt.Fprintf(buf, "\t%s -> %s\n", cond, c)
	}
	fmt.Fprintf(buf, "\t%s -> %s\n", cond, defaultLabel)
	for _, c := range cases {
		fmt.Fprintf(buf, "\t%s -> %s\n", c, exit)
	}
	fmt.Fprintf(buf, "\t%s -> %s\n", defaultLabel, exit)
	buf.WriteString("}")
	return buf.String()
}

// FindSwitch returns the first occurrence of an n-way conditional statement in
// g, and a boolean indicating if such a primitive was found.
func FindSwitch(g graph.Directed, dom cfa.DominatorTree) (prim Switch, ok bool) {
	// Range through cond node candidates.
	for nodes := g.Nodes(); nodes.Next(); {
		// Note: This run-time type assertion goes away, should Gonum graph start
		// to leverage generics in Go2.
		cond := nodes.Node().(cfa.Node)
		// Verify that cond has at least one successor (default).
		condSuccs := cfa.NodesOf(g.From(cond.ID()))
		if len(condSuccs) < 1 {
			continue
		}
		prim.Cond = cond
		// Select cases and default node candidates.
		// TODO: try each combination for default node.
		prim.Cases = condSuccs[:len(condSuccs)-1]
		prim.Default = condSuccs[len(condSuccs)-1]
		// Verify that default has one successor (exit).
		defaultSuccs := cfa.NodesOf(g.From(prim.Default.ID()))
		if len(defaultSuccs) != 1 {
			continue
		}
		// Select exit node candidate.
		prim.Exit = defaultSuccs[0]
		if prim.IsValid(g, dom) {
			return prim, true
		}
	}
	return Switch{}, false
}

// IsValid reports whether the cond, case_B, case_C, case_D, ..., default and
// exit node candidates of prim form a valid n-way conditional statement in g.
//
// Control flow graph:
//
//              cond
//         ↙      ↓       ↘      ↘       ↘
//    case_B   case_C   case_D   ...   default
//        ↘       ↓       ↙      ↙       ↙
//              exit
func (prim Switch) IsValid(g graph.Directed, dom cfa.DominatorTree) bool {
	// Dominator sanity check.
	cond, cases, defaultNode, exit := prim.Cond, prim.Cases, prim.Default, prim.Exit
	for _, c := range cases {
		if !dom.Dominates(cond, c) {
			return false
		}
	}
	if !dom.Dominates(cond, defaultNode) || !dom.Dominates(cond, exit) {
		return false
	}
	// Verify that cond has n successors (where n = len(cases) + 1).
	condSuccs := g.From(cond.ID())
	if condSuccs.Len() != len(cases)+1 {
		return false
	}
	for _, c := range cases {
		if !g.HasEdgeFromTo(cond.ID(), c.ID()) {
			return false
		}
	}
	if !g.HasEdgeFromTo(cond.ID(), defaultNode.ID()) {
		return false
	}
	// Verify that each case node has one predecessor (cond) and one successor (exit).
	for _, c := range cases {
		caseSuccs := g.From(c.ID())
		casePreds := g.To(c.ID())
		if casePreds.Len() != 1 || caseSuccs.Len() != 1 || !g.HasEdgeFromTo(c.ID(), exit.ID()) {
			return false
		}
	}
	// Verify that default has one predecessor (cond) and one successor (exit).
	defaultSuccs := g.From(defaultNode.ID())
	defaultPreds := g.To(defaultNode.ID())
	if defaultPreds.Len() != 1 || defaultSuccs.Len() != 1 || !g.HasEdgeFromTo(defaultNode.ID(), exit.ID()) {
		return false
	}
	// Verify that exit has n predecessor (where n = len(cases) + 1).
	exitPreds := g.To(exit.ID())
	return exitPreds.Len() == len(cases)+1
}
