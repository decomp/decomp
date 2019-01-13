package interval

import (
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
)

// Analyze analyzes the given control flow graph and returns the list of
// recovered high-level control flow primitives. The before and after functions
// are invoked if non-nil before and after merging the nodes of located
// primitives.
func Analyze(g cfa.Graph, before, after func(g cfa.Graph, prim *primitive.Primitive)) []*primitive.Primitive {
	var prims []*primitive.Primitive
	// Initialize depth-first search visit order.
	initDFSOrder(g)

	// The Structuring Algorithm is not finite Church-Rosser. Thus an ordering is
	// to be followed, namely: structure n-way conditionals, followed by loop
	// structuring, and 2-way conditional structuring last.

	// Structure compound conditionals.
	prims = append(prims, structCompCond(g)...)
	// Structure n-way conditionals.
	dom := cfa.NewDom(g)
	prims = append(prims, structNway(g, dom)...)
	// Structure loops of the control flow graph.
	prims = append(prims, loopStruct(g)...)
	// Structure 2-way conditionals.
	prims = append(prims, struct2way(g, dom)...)
	return prims
}
