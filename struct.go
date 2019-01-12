package interval

import (
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
)

// Analyze analyzes the given control flow graph and returns the list of
// recovered high-level control flow primitives. The before and after functions
// are invoked if non-nil before and after merging the nodes of located
// primitives.
func Analyze(g cfa.Graph, before, after func(g cfa.Graph, prim *primitive.Primitive)) ([]*primitive.Primitive, error) {
	var prims []*primitive.Primitive
	// Initialize depth-first search visit order.
	initDFSOrder(g)

	// The Structuring Algorithm is not finite Church-Rosser. Thus an ordering is
	// to be followed, namely: structure n-way conditionals, followed by loop
	// structuring, and 2-way conditional structuring last.

	// Structure n-way conditionals.
	dom := cfa.NewDom(g)
	structNway(g, dom)
	// Structure loops of the control flow graph.
	loopStruct(g)
	// Structure 2-way conditionals.
	struct2way(g, dom)
	// Structure compound conditionals.
	structCompCond(g)

	// TODO: compute recovered primitives (prims) from structuring information.

	return prims, nil
}
