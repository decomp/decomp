// Package hammock implements the Hammock method control flow recovery
// algorithm.
//
// At a high-level, the Hammock method identifies subgraph isomorphisms of the
// cannonical subgraph representations of high-level control flow primitives in
// control flow graphs. Once a subgraph isomorphism has been located, the nodes
// of the subgraph are collapsed into a single node, and the process is
// repreated, either until the control flow graph contains a single node, or no
// more subgraph isomorphisms may be located.
package hammock

import (
	goerrors "errors"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"github.com/pkg/errors"
)

// ErrIncomplete signals an incomplete control flow recovery.
var ErrIncomplete = goerrors.New("incomplete control flow recovery")

// Analyze analyzes the given control flow graph and returns the list of
// recovered high-level control flow primitives. The before and after functions
// are invoked if non-nil before and after merging the nodes of located
// primitives.
func Analyze(g cfa.Graph, before, after func(g cfa.Graph, prim *primitive.Primitive)) ([]*primitive.Primitive, error) {
	var prims []*primitive.Primitive
	for {
		// Locate control flow primitive.
		dom := cfa.NewDom(g)
		prim, ok := FindPrim(g, dom)
		if !ok {
			break
		}
		prims = append(prims, prim)
		if before != nil {
			before(g, prim)
		}
		// Merge nodes of located primitive.
		newG, err := cfa.Merge(g, prim)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		g = newG
		if after != nil {
			after(g, prim)
		}
	}
	if g.Nodes().Len() > 1 {
		// Return partial results and signal incomplete control flow recovery.
		return prims, ErrIncomplete
	}
	return prims, nil
}

// FindPrim returns the first occurrence of a high-level control flow primitive
// in g, and a boolean indicating if such a primitive was found.
func FindPrim(g cfa.Graph, dom cfa.DominatorTree) (*primitive.Primitive, bool) {
	// Locate sequences of two statements.
	if prim, ok := FindSeq(g, dom); ok {
		return prim.Prim(), true
	}
	// Locate pre-test loops.
	if prim, ok := FindPreLoop(g, dom); ok {
		return prim.Prim(), true
	}
	// Locate post-test loops.
	if prim, ok := FindPostLoop(g, dom); ok {
		return prim.Prim(), true
	}
	// Locate 1-way conditionals.
	if prim, ok := FindIf(g, dom); ok {
		return prim.Prim(), true
	}
	// Locate 1-way conditionals with a body return statements.
	if prim, ok := FindIfReturn(g, dom); ok {
		return prim.Prim(), true
	}
	// Locate 2-way conditionals.
	if prim, ok := FindIfElse(g, dom); ok {
		return prim.Prim(), true
	}
	// TODO: Locate n-way conditionals.
	//if prim, ok := FindSwitch(g, dom); ok {
	//	return prim.Prim(), true
	//}
	return nil, false
}
