package interval

import (
	"fmt"
	"math"

	"github.com/mewmew/lnp/pkg/cfa"
)

// loopStruct structures loops in the given control flow graph.
//
// Pre: G^1 ... G^n has been constructed.
//      II^1 ... II^n has been determined.
//
// Post: All nodes of G that belong to a loop are marked. All loop header nodes
//       have information on the type of loop and the latching node.
//
// ref: Figure 6-25; Cifuentes' Reverse Comilation Techniques.
func loopStruct(g cfa.Graph) {
	Gs, IIs := DerivedSequence(g)
	if len(Gs) != len(IIs) {
		panic(fmt.Errorf("length mismatch between derived sequence of graphs (%d) and corresponding intervals (%d)", len(Gs), len(IIs)))
	}
	// for (G^i := G^1 ... G^n)
	for i, Gi := range Gs {
		// for (I^i(h_j)) := I^1(h_1) ... I^m(h_m)
		for j, I := range IIs[i] {
			_ = j
			// if ((\exists x \in N^i, (x, h_j) \in E^i) \land (inLoop(x) == False))
			for xs := Gi.To(I.head.ID()); xs.Next(); {
				x := xs.Node().(*Node)
				if !x.inLoop {
					// for (all n \in loop (x, h_j))
					nodesInLoop := loop(I, x)
					for _, n := range nodesInLoop {
						// inLoop(n) = True
						n.inLoop = true
					}
					// loopType(h_j) = findLoopType((x, h_j))
					I.head.loopType = findLoopType(I, x, nodesInLoop)
					// loopFollow(h_j) = findLoopFollow((x, h_j))
					I.head.loopFollow = findLoopFollow(I, x, nodesInLoop)
					break
				}
			}
		}
	}
}

// TODO: implement and use markNodesInLoop instead of loop. Get rid of n.inLoop
// in favour of loopHead?
//
// Pre: (latch, I.head) is a back-edge.
//
// Post: the nodes that belong to the loop (latch, I.head) are marked.
//
// ref: Figure 6-27; Cifuentes' Reverse Comilation Techniques.

// loop returns the nodes of the loop (latch, I.head).
func loop(I *Interval, latch *Node) []*Node {
	// The nodes belong to the same interval, since the interval header (i.e. x)
	// dominates all nodes of the interval, and in a loop, the loop header node
	// dominates all nodes of the loop. If a node belongs to a different
	// interval, it is not dominated by the loop header node, thus it cannot
	// belong to the same loop.
	//
	//    \forall n \in loop(y, x), n \in I(x)
	var ns []*Node
	for nodes := I.Nodes(); nodes.Next(); {
		n := nodes.Node().(*Node)
		// The loop is formed of all nodes that are between x and y in terms of
		// node numbering.
		//
		//    \forall n \in loop(y, x), n \in {x ... y}
		if I.head.postNum <= n.postNum && n.postNum <= latch.postNum {
			ns = append(ns, n)
		}
	}
	return ns
}

// loopType is the set of loop types.
type loopType uint8

// Loop types.
const (
	loopTypeNone loopType = iota
	// Pre-test loop.
	loopTypePreTest
	// Post-test loop.
	loopTypePostTest
	// Endless loop.
	loopTypeEndless
)

// findLoopType returns the type of the loop (latch, I.head).
//
// Pre: (latch, I.head) induces a loop.
//
//      nodesInLoop is the set of all nodes that belong to the loop (latch,
//      I.head).
//
// Post: loopType(I.head) has the type of loop induces by (latch, I.head).
//
// ref: Figure 6-28; Cifuentes' Reverse Comilation Techniques.
func findLoopType(I *Interval, latch *Node, nodesInLoop []*Node) loopType {
	headSuccs := NodesOf(I.From(I.head.ID()))
	latchSuccs := NodesOf(I.From(latch.ID()))
	switch len(latchSuccs) {
	// if (nodeType(y) == 2-way)
	case 2:
		switch len(headSuccs) {
		// if (nodeType(x) == 2-way)
		case 2:
			// if (outEdge(x, 1) \in nodesInLoop \land (outEdge(x, 2) \in nodesInLoop)
			if contains(nodesInLoop, headSuccs[0]) && contains(nodesInLoop, headSuccs[1]) {
				// loopType(x) = Post_Tested.
				return loopTypePostTest
			} else {
				// loopType(x) = Pre_Tested.
				return loopTypePreTest
			}
		// 1-way header node.
		case 1:
			// loopType(x) = Post_Tested.
			return loopTypePostTest
		default:
			panic(fmt.Errorf("support for %d-way header node not yet implemented", len(headSuccs)))
		}
	// 1-way latching node.
	case 1:
		switch len(headSuccs) {
		// if nodeType(x) == 2-way
		case 2:
			// loopType(x) = Pre_Tested.
			return loopTypePreTest
		// 1-way header node.
		case 1:
			// loopType(x) = Endless.
			return loopTypeEndless
		default:
			panic(fmt.Errorf("support for %d-way header node not yet implemented", len(headSuccs)))
		}
	default:
		panic(fmt.Errorf("support for %d-way latch node not yet implemented", len(latchSuccs)))
	}
}

// findLoopFollow returns the follow node of the loop (latch, I.head).
//
// Pre: (latch, I.head) induces a loop.
//
//      nodesInLoop is the set of all nodes that belong to the loop (latch,
//      I.head).
//
// Post: loopFollow(latch) is the follow node to the loop induces by (latch,
//       I.head).
//
// ref: Figure 6-29; Cifuentes' Reverse Comilation Techniques.
func findLoopFollow(I *Interval, latch *Node, nodesInLoop []*Node) *Node {
	headSuccs := NodesOf(I.From(I.head.ID()))
	latchSuccs := NodesOf(I.From(latch.ID()))
	switch I.head.loopType {
	// if (loopType(x) == Pre_Tested)
	case loopTypePreTest:
		switch {
		// if (outEdges(x, 1) \in nodesInLoop)
		case contains(nodesInLoop, headSuccs[0]):
			// loopFollow(x) = outEdges(x, 2)
			return headSuccs[1]
		case contains(nodesInLoop, headSuccs[1]):
			// loopFollow(x) = outEdges(x, 1)
			return headSuccs[0]
		default:
			panic(fmt.Errorf("unable to locate follow loop of pre-test header node %q", I.head.DOTID()))
		}
	// else if (loopType(x) == Post_Tested)
	case loopTypePostTest:
		switch {
		// if (outEdges(y, 1) \in nodesInLoop)
		case contains(nodesInLoop, latchSuccs[0]):
			// loopFollow(x) = outEdges(y, 2)
			return latchSuccs[1]
		case contains(nodesInLoop, latchSuccs[1]):
			// loopFollow(x) = outEdges(y, 1)
			return latchSuccs[0]
		default:
			panic(fmt.Errorf("unable to locate follow loop of post-test latch node %q", latch.DOTID()))
		}
	// endless loop.
	case loopTypeEndless:
		// fol = Max // a large constant.
		followPostNum := math.MaxInt64
		var follow *Node
		// for (all 2-way nodes n \in nodesInLoop)
		for _, n := range nodesInLoop {
			nSuccs := NodesOf(I.From(n.ID()))
			if len(nSuccs) != 2 {
				// Skip node as not 2-way conditional.
				continue
			}
			switch {
			// if ((outEdges(n, 1) \not \in nodesInLoop) \land (outEdges(x, 1) < fol))
			case !contains(nodesInLoop, nSuccs[0]) && nSuccs[0].postNum < followPostNum:
				followPostNum = nSuccs[0].postNum
				follow = nSuccs[0]
			// if ((outEdges(x, 2) \not \in nodesInLoop) \land (outEdges(x, 2) < fol))			}
			case !contains(nodesInLoop, nSuccs[1]) && nSuccs[1].postNum < followPostNum:
				followPostNum = nSuccs[1].postNum
				follow = nSuccs[1]
			}
		}
		// if (fol != Max)
		if followPostNum != math.MaxInt64 {
			// loopFollow(x) = fol
			return follow
		}
		// No follow node located.
		return nil
	default:
		panic(fmt.Errorf("support for loop type %v not yet implemented", I.head.loopType))
	}
}

// contains reports whether the list of nodes contains the given node.
func contains(ns []*Node, n *Node) bool {
	for i := range ns {
		if ns[i].ID() == n.ID() {
			return true
		}
	}
	return false
}
