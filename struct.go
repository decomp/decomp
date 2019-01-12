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
	// Structure loops of the control flow graph.
	loopStruct(g)
	// Structure 2-way conditionals.
	dom := cfa.NewDom(g)
	struct2way(g, dom)
	// TODO: implement structuring of 2-way, n-way and short-circuit.
	return prims, nil
}
