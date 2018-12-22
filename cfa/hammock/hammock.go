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
	"fmt"

	"github.com/mewmew/lnp/cfa"
	"github.com/mewmew/lnp/cfa/primitive"
	"github.com/pkg/errors"
)

// Analyze analyzes the given control flow graph and returns the list of
// recovered high-level control flow primitives.
func Analyze(g cfa.Graph) ([]*primitive.Primitive, error) {
	var prims []*primitive.Primitive
	for {
		// Locate control flow primitive.
		dom := cfa.NewDom(g)
		prim, ok := FindPrim(g, dom)
		if !ok {
			break
		}
		prims = append(prims, prim)
		fmt.Println("before merge:", g)
		// Merge nodes of located primitive.
		fmt.Println("located prim:", prim)
		newG, err := cfa.Merge(g, prim)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		fmt.Println("after merge:", newG)
		g = newG
	}
	return prims, nil
}

// FindPrim returns the first occurrence of a high-level control flow primitive
// in g, and a boolean indicating if such a primitive was found.
func FindPrim(g cfa.Graph, dom cfa.DominatorTree) (*primitive.Primitive, bool) {
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
	// Locate sequences of two statements.
	if prim, ok := FindSeq(g, dom); ok {
		return prim.Prim(), true
	}
	return nil, false
}
