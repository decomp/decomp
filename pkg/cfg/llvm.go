package cfg

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/mewmew/lnp/pkg/cfa"
	"gonum.org/v1/gonum/graph/encoding"
)

// NewGraphFromFunc returns a new control flow graph based on the given LLVM IR
// function.
func NewGraphFromFunc(f *ir.Func) *Graph {
	g := NewGraph()
	// Force generate local IDs.
	if err := f.AssignIDs(); err != nil {
		panic(fmt.Errorf("unable to assign IDs to locate variables of function %q; %v", f.Ident(), err))
	}
	// Generate control flow graph of function.
	for i, block := range f.Blocks {
		from := nodeWithName(g, block.Name())
		if i == 0 {
			// Set entry basic block.
			g.SetEntry(from)
		}
		switch term := block.Term.(type) {
		case *ir.TermRet:
			// TODO: consider adding attribute to distinguish return instructions
			// in CFG.

			// nothing to do.
		case *ir.TermBr:
			to := nodeWithName(g, term.Target.Name())
			edgeWithLabel(g, from, to, "")
		case *ir.TermCondBr:
			targetTrue := nodeWithName(g, term.TargetTrue.Name())
			targetFalse := nodeWithName(g, term.TargetFalse.Name())
			edgeWithLabel(g, from, targetTrue, "true")
			edgeWithLabel(g, from, targetFalse, "false")
		case *ir.TermSwitch:
			for _, c := range term.Cases {
				to := nodeWithName(g, c.Target.Name())
				label := fmt.Sprintf("case (%v=%v)", term.X.Ident(), c.X.Ident())
				edgeWithLabel(g, from, to, label)
			}
			to := nodeWithName(g, term.TargetDefault.Name())
			edgeWithLabel(g, from, to, "default case")
		//case *ir.TermIndirectBr:
		//case *ir.TermInvoke:
		//case *ir.TermResume:
		//case *ir.TermCatchSwitch:
		//case *ir.TermCatchRet:
		//case *ir.TermCleanupRet:
		case *ir.TermUnreachable:
			// nothing to do.
		default:
			panic(fmt.Errorf("support for terminator %T not yet implemented", term))
		}
	}
	return g
}

// ### [ Helper functions ] ####################################################

// edgeWithLabel adds a directed edge between the specified nodes and assignes
// it the given label.
func edgeWithLabel(g cfa.Graph, from, to cfa.Node, label string) cfa.Edge {
	e := g.NewEdge(from, to).(cfa.Edge)
	if len(label) > 0 {
		// Skip label for true and false, just colour edge.
		switch label {
		case "true":
			e.SetAttribute(encoding.Attribute{Key: "color", Value: "darkgreen"})
		case "false":
			e.SetAttribute(encoding.Attribute{Key: "color", Value: "red"})
		default:
			e.SetAttribute(encoding.Attribute{Key: "label", Value: label})
		}
		e.SetAttribute(encoding.Attribute{Key: "cond", Value: label})
	}
	g.SetEdge(e)
	return e
}

// nodeWithName returns the node of the given name. A new node is created if not
// yet present in the control flow graph.
func nodeWithName(g cfa.Graph, name string) cfa.Node {
	if n, ok := g.NodeWithDOTID(name); ok {
		return n
	}
	n := g.NewNode().(cfa.Node)
	n.SetDOTID(name)
	g.AddNode(n)
	return n
}
