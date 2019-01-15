package interval

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
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
func loopStruct(g cfa.Graph, dom cfa.DominatorTree) []*primitive.Primitive {
	var prims []*primitive.Primitive
	Gs, IIs := DerivedSequence(g)
	if len(Gs) != len(IIs) {
		panic(fmt.Errorf("length mismatch between derived sequence of graphs (%d) and corresponding intervals (%d)", len(Gs), len(IIs)))
	}
	// for (G^i := G^1 ... G^n)
	for i := range Gs {
		// for (I^i(h_j)) := I^1(h_1) ... I^m(h_m)
		for _, I := range IIs[i] {
			// if ((\exists x \in N^i, (x, h_j) \in E^i) \land (inLoop(x) == False))
			if head, latch, ok := findLatch(g, I, IIs); ok && latch.LoopHead == nil {
				// Mark node as latch node (to not be used in 2-way conditions).
				latch.IsLoopLatch = true
				// for (all n \in loop (x, h_j))
				nodesInLoop := markNodesInLoop(g, head, latch, dom)
				// loopType(h_j) = findLoopType((x, h_j))
				head.LoopType = findLoopType(g, head, latch, nodesInLoop)
				// loopFollow(h_j) = findLoopFollow((x, h_j))
				head.LoopFollow = findLoopFollow(g, head, latch, nodesInLoop)
				// Create primitive.
				prim := &primitive.Primitive{
					Prim:  head.LoopType.String(), // pre_loop, post_loop or inf_loop
					Entry: head.DOTID(),
					Nodes: map[string]string{
						"latch": latch.DOTID(),
					},
				}
				// Add loop body nodes to primitive.
				for i, n := range nodesInLoop {
					name := fmt.Sprintf("body_%d", i)
					prim.Nodes[name] = n.DOTID()
				}
				prims = append(prims, prim)
			}
		}
	}
	return prims
}

// findLatch locates the loop latch node in the interval, based on the interval
// header node. The boolean return value indicates success.
func findLatch(g cfa.Graph, I *Interval, IIs [][]*Interval) (head, latch *Node, ok bool) {
	// iis is used to look up the nodes belonging to an interval, e.g. I_1. Note,
	// iis is 0-indexed.
	var iis []*Interval
	for _, Is := range IIs {
		iis = append(iis, Is...)
	}
	// Each header of an interval in G^i is checked for having a back-edge from a
	// latching node that belong to the same interval.
	for preds := I.To(I.head.ID()); preds.Next(); {
		pred := preds.Node().(*Node)
		if latch == nil || pred.RevPostNum > latch.RevPostNum {
			latch = pred
		}
	}
	if latch != nil {
		// Locate node in original control flow graph corresponding to the latch
		// node in the derived sequence of graphs.
		if l, ok := g.NodeWithDOTID(latch.DOTID()); ok {
			return I.head, l.(*Node), true
		}
		h := findOrigHead(iis, I.head)
		latchCands := descRevPostOrder(NodesOf(g.To(h.ID())))
		for i, latchCand := range latchCands {
			if latchCand.RevPostNum < h.RevPostNum {
				latchCands = latchCands[:i]
				break
			}
		}
		l := findOrigLatch(iis, latchCands, latch)
		return h, l, true
	}
	return nil, nil, false
}

// findOrigHead returns the loop header node in the original control flow graph
// corresponding to the header node of an interval in the derived sequence of
// graphs.
func findOrigHead(iis []*Interval, head *Node) *Node {
	// Find the outer-most interval which has the loop header as interval header.
	i, ok := getInterval(iis, head.DOTID())
	if !ok {
		return head
	}
	return findOrigHead(iis, i.head)
}

// findOrigLatch returns the latch node in the original control flow graph
// corresponding to the latch node of an interval in the derived sequence of
// graphs.
func findOrigLatch(iis []*Interval, latchCands []*Node, latch *Node) *Node {
	i, ok := getInterval(iis, latch.DOTID())
	if !ok {
		return latch
	}
	l, ok := findNodeInInterval(iis, i, latchCands)
	if !ok {
		panic(fmt.Errorf("unable to locate original latch node %q", latch.DOTID()))
	}
	return l
}

// findNodeInInterval locates the a latch node in the original control flow
// graph corresponding to one of the latch node candidates in the derived
// sequence of graphs.
func findNodeInInterval(iis []*Interval, i *Interval, latchCands []*Node) (*Node, bool) {
	for _, latchCand := range latchCands {
		for nodes := i.Nodes(); nodes.Next(); {
			n := nodes.Node().(*Node)
			j, ok := getInterval(iis, n.DOTID())
			if !ok {
				if n.DOTID() == latchCand.DOTID() {
					return n, true
				}
			} else if l, ok := findNodeInInterval(iis, j, latchCands); ok {
				return l, true
			}
		}
	}
	return nil, false
}

// getInterval returns the interval of the given node (with DOT ID e.g. "I_42").
// The boolean return value indicates success.
func getInterval(iis []*Interval, dotID string) (*Interval, bool) {
	if !strings.HasPrefix(dotID, "I_") {
		return nil, false
	}
	dotID = dotID[len("I_"):]
	hid, err := strconv.Atoi(dotID)
	if err != nil {
		panic(fmt.Errorf("unable to parse interval node ID %q; %v", dotID, err))
	}
	i := iis[hid-1]
	return i, true
}

// loop returns the nodes of the loop (latch, I.head), marking the loop header
// of each node.
//
// Pre: (latch, I.head) is a back-edge.
//
// Post: the nodes that belong to the loop (latch, I.head) are marked.
//
// ref: Figure 6-27; Cifuentes' Reverse Comilation Techniques.
func markNodesInLoop(g cfa.Graph, head, latch *Node, dom cfa.DominatorTree) []*Node {
	nodesInLoop := []*Node{head}
	head.LoopHead = head
	for _, n := range ascRevPostOrder(NodesOf(g.Nodes())) {
		// The loop is formed of all nodes that are between x and y in terms of
		// node numbering.
		if head.RevPostNum < n.RevPostNum && n.RevPostNum <= latch.RevPostNum {
			// The nodes belong to the same interval, since the interval header
			// (i.e. x) dominates all nodes of the interval, and in a loop, the
			// loop header node dominates all nodes of the loop. If a node belongs
			// to a different interval, it is not dominated by the loop header
			// node, thus it cannot belong to the same loop.
			if dom.Dominates(head.ID(), n.ID()) {
				nodesInLoop = append(nodesInLoop, n)
				if n.LoopHead == nil {
					n.LoopHead = head
				}
			}
		}
		if n.RevPostNum > latch.RevPostNum {
			break
		}
	}
	return nodesInLoop
}

//go:generate stringer -linecomment -type LoopType

// LoopType is the set of loop types.
type LoopType uint8

// Loop types.
const (
	LoopTypeNone LoopType = iota
	// Pre-test loop.
	LoopTypePreTest // pre_loop
	// Post-test loop.
	LoopTypePostTest // post_loop
	// Endless loop.
	LoopTypeEndless // inf_loop
)

// findLoopType returns the type of the loop (latch, head).
//
// Pre: (latch, head) induces a loop.
//
//      nodesInLoop is the set of all nodes that belong to the loop (latch,
//      head).
//
// Post: loopType(head) has the type of loop induces by (latch, head).
//
// ref: Figure 6-28; Cifuentes' Reverse Comilation Techniques.
func findLoopType(g cfa.Graph, head, latch *Node, nodesInLoop []*Node) LoopType {
	// Add extra case not present in Cifuentes' for when head == latch.
	if head.ID() == latch.ID() {
		return LoopTypePostTest
	}
	headSuccs := NodesOf(g.From(head.ID()))
	latchSuccs := NodesOf(g.From(latch.ID()))
	switch len(latchSuccs) {
	// if (nodeType(y) == 2-way)
	case 2:
		switch len(headSuccs) {
		// if (nodeType(x) == 2-way)
		case 2:
			// if (outEdge(x, 1) \in nodesInLoop \land (outEdge(x, 2) \in nodesInLoop)
			if contains(nodesInLoop, headSuccs[0]) && contains(nodesInLoop, headSuccs[1]) {
				// loopType(x) = Post_Tested.
				return LoopTypePostTest
			} else {
				// loopType(x) = Pre_Tested.
				return LoopTypePreTest
			}
		// 1-way header node.
		case 1:
			// loopType(x) = Post_Tested.
			return LoopTypePostTest
		default:
			panic(fmt.Errorf("support for %d-way header node not yet implemented", len(headSuccs)))
		}
	// 1-way latching node.
	case 1:
		switch len(headSuccs) {
		// if nodeType(x) == 2-way
		case 2:
			// loopType(x) = Pre_Tested.
			return LoopTypePreTest
		// 1-way header node.
		case 1:
			// loopType(x) = Endless.
			return LoopTypeEndless
		default:
			panic(fmt.Errorf("support for %d-way header node not yet implemented", len(headSuccs)))
		}
	default:
		panic(fmt.Errorf("support for %d-way latch node not yet implemented", len(latchSuccs)))
	}
}

// findLoopFollow returns the follow node of the loop (latch, head).
//
// Pre: (latch, head) induces a loop.
//
//      nodesInLoop is the set of all nodes that belong to the loop (latch,
//      head).
//
// Post: loopFollow(latch) is the follow node to the loop induces by (latch,
//       head).
//
// ref: Figure 6-29; Cifuentes' Reverse Comilation Techniques.
func findLoopFollow(g cfa.Graph, head, latch *Node, nodesInLoop []*Node) *Node {
	headSuccs := NodesOf(g.From(head.ID()))
	latchSuccs := NodesOf(g.From(latch.ID()))
	switch head.LoopType {
	// if (loopType(x) == Pre_Tested)
	case LoopTypePreTest:
		switch {
		// if (outEdges(x, 1) \in nodesInLoop)
		case contains(nodesInLoop, headSuccs[0]):
			// loopFollow(x) = outEdges(x, 2)
			return headSuccs[1]
		case contains(nodesInLoop, headSuccs[1]):
			// loopFollow(x) = outEdges(x, 1)
			return headSuccs[0]
		default:
			panic(fmt.Errorf("unable to locate follow loop of pre-test header node %q", head.DOTID()))
		}
	// else if (loopType(x) == Post_Tested)
	case LoopTypePostTest:
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
	case LoopTypeEndless:
		// fol = Max // a large constant.
		followRevPostNum := math.MaxInt64
		var follow *Node
		// for (all 2-way nodes n \in nodesInLoop)
		for _, n := range nodesInLoop {
			nSuccs := NodesOf(g.From(n.ID()))
			if len(nSuccs) != 2 {
				// Skip node as not 2-way conditional.
				continue
			}
			switch {
			// if ((outEdges(n, 1) \not \in nodesInLoop) \land (outEdges(x, 1) < fol))
			case !contains(nodesInLoop, nSuccs[0]) && nSuccs[0].RevPostNum < followRevPostNum:
				followRevPostNum = nSuccs[0].RevPostNum
				follow = nSuccs[0]
			// if ((outEdges(x, 2) \not \in nodesInLoop) \land (outEdges(x, 2) < fol))			}
			case !contains(nodesInLoop, nSuccs[1]) && nSuccs[1].RevPostNum < followRevPostNum:
				followRevPostNum = nSuccs[1].RevPostNum
				follow = nSuccs[1]
			}
		}
		// if (fol != Max)
		if followRevPostNum != math.MaxInt64 {
			// loopFollow(x) = fol
			return follow
		}
		// No follow node located.
		return nil
	default:
		panic(fmt.Errorf("support for loop type %v not yet implemented", head.LoopType))
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
