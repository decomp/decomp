// Package cfa implements control flow analysis of control flow graphs.
package cfa

import (
	"fmt"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"github.com/gonum/graph"
	"github.com/pkg/errors"
)

// FindPrim locates a control flow primitive in the provided control flow graph
// and merges its nodes into a single node.
func FindPrim(g graph.Directed, dom cfg.Dom) (*primitive.Primitive, error) {
	// Locate pre-test loops.
	if prim, ok := FindPreLoop(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate post-test loops.
	if prim, ok := FindPostLoop(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate 1-way conditionals.
	if prim, ok := FindIf(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate 1-way conditionals with a body return statements.
	if prim, ok := FindIfReturn(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate 2-way conditionals.
	if prim, ok := FindIfElse(g, dom); ok {
		return prim.Prim(), nil
	}

	// Locate sequences of two statements.
	if prim, ok := FindSeq(g, dom); ok {
		return prim.Prim(), nil
	}

	return nil, errors.Errorf("unable to locate control flow primitive")
}

// label returns the label of the node.
func label(n graph.Node) string {
	if n, ok := n.(*cfg.Node); ok {
		return n.Label
	}
	panic(fmt.Sprintf("invalid node type; expected *cfg.Node, got %T", n))
}
