// Package interval implements the Interval method control flow recovery
// algorithm.
//
// At a high-level, the Interval method TODO...
package interval

import (
	"log"
	"os"

	"github.com/mewkiz/pkg/term"
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
)

var (
	// dbg represents a logger with the "interval:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.YellowBold("interval:")+" ", 0)
	// warn represents a logger with the "interval:" prefix, which logs warning
	// messages to standard error.
	warn = log.New(os.Stderr, term.RedBold("interval:")+" ", 0)
)

// Analyze analyzes the given control flow graph and returns the list of
// recovered high-level control flow primitives. The before and after functions
// are invoked if non-nil before and after merging the nodes of located
// primitives.
func Analyze(g cfa.Graph, before, after func(g cfa.Graph, prim *primitive.Primitive)) ([]*primitive.Primitive, error) {
	panic("not yet implemented")
}
