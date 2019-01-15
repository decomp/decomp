package interval

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"github.com/rickypai/natsort"
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
		fmt.Printf("G_%d\n", i)
		// for (I^i(h_j)) := I^1(h_1) ... I^m(h_m)
		for _, I := range IIs[i] {
			// if ((\exists x \in N^i, (x, h_j) \in E^i) \land (inLoop(x) == False))
			if head, latch, ok := findLatch(g, I, IIs); ok && latch.LoopHead == nil {
				fmt.Println("head:", head.DOTID())
				fmt.Println("latch:", latch.DOTID())
				// Mark node as latch node (to not be used in 2-way conditions).
				latch.IsLoopLatch = true
				// for (all n \in loop (x, h_j))
				//fmt.Println("=== [ loop nodes ] ===") // TODO: remove debug output.
				nodesInLoop := markNodesInLoop(g, head, latch, dom)
				//printNodes(nodesInLoop) // TODO: remove debug output
				// loopType(h_j) = findLoopType((x, h_j))
				head.LoopType = findLoopType(g, head, latch, nodesInLoop)
				// loopFollow(h_j) = findLoopFollow((x, h_j))
				head.LoopFollow = findLoopFollow(g, head, latch, nodesInLoop)
				// Create primitive.
				prim := &primitive.Primitive{
					Prim:  head.LoopType.String(), // pre_loop, post_loop or inf_loop
					Entry: head.DOTID(),
					Nodes: map[string]string{
						// TODO: Include entry node?
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
	fmt.Println("=== [ findLatch ] ===")
	fmt.Println("   head:", I.head.DOTID())
	for preds := I.To(I.head.ID()); preds.Next(); {
		pred := preds.Node().(*Node)
		if latch == nil || pred.RevPostNum > latch.RevPostNum {
			latch = pred
		}
	}
	if latch != nil {
		fmt.Println("   candidate latch:", latch.DOTID())
		// Find original latch node.
		if l, ok := g.NodeWithDOTID(latch.DOTID()); ok {
			fmt.Println("   orig latch:", l.DOTID())
			return I.head, l.(*Node), true
		}
		h := findOrigHead(iis, I.head)
		fmt.Println("   orig head:", h.DOTID())
		latchCands := descRevPostOrder(NodesOf(g.To(h.ID())))
		for i, latchCand := range latchCands {
			if latchCand.RevPostNum < h.RevPostNum {
				latchCands = latchCands[:i]
				break
			}
		}
		fmt.Println("   latch candidates:")
		printNodes(latchCands)
		l := findOrigLatch(iis, latchCands, latch)
		fmt.Println("   orig latch located:", l.DOTID())
		return h, l, true
	}
	return nil, nil, false

	/*
		I := IIs[i][j]
		// iis is used to look up the nodes belonging to an interval, e.g. I_1. Note,
		// iis is 0-indexed.
		var iis []*Interval
		for _, Is := range IIs {
			iis = append(iis, Is...)
		}
		head := I.head
		fmt.Println("=== [ find latch ] ===")
		fmt.Println("head:", head.DOTID())
		fmt.Println(I.String())
		fmt.Println()
		// Each header of an interval in G^i is checked for having a back-edge from a
		// latching node that belong to the same interval.
		for preds := I.To(head.ID()); preds.Next(); {
			x := preds.Node().(*Node)
			// Find back-edge.
			// Latch node located.
			fmt.Println("   ==> candidate latch located:", x.DOTID())
			if latch == nil || x.RevPostNum > latch.RevPostNum {
				latch = x
			}
		}
		h := findOrigHead(iis, head)
		// TODO: use h hence forth.
		// Latch node candidates in original graph.
		fmt.Println("orig head:", h.DOTID())
		latchCands := descRevPostOrder(NodesOf(g.To(h.ID())))
		fmt.Printf("~~~~~~~ [[ latch candidates of %q ]] ~~~~~~~~~~~~~~~~~~~\n", h.DOTID())
		printNodes(latchCands)
		fmt.Println("~~~~~~~ [[/ latch candidates ]] ~~~~~~~~~~~~~~~~~~~")
		// Skip candidates with reverse post-order number less than loop header.
		l := 0
		for _, latchCand := range latchCands {
			if latchCand.RevPostNum < h.RevPostNum {
				break
			}
			l++
		}
		latchCands = latchCands[:l]
		fmt.Printf("~~~~~~~>> [[ latch candidates 2 of %q ]] ~~~~~~~~~~~~~~~~~~~\n", h.DOTID())
		fmt.Println("h rev post:", h.RevPostNum)
		printNodes(latchCands)
		fmt.Println("~~~~~~~>> [[/ latch candidates 2 ]] ~~~~~~~~~~~~~~~~~~~")
		if latch != nil {
			// Locate node in original control flow graph corresponding to the latch
			// node in the derived sequence of graphs.
			headID := h.ID()
			// Find the outer-most interval which has the loop header as interval
			// header.
			for ii := i; ii >= 0; ii-- {
				for _, J := range IIs[ii] {
					if J.head.ID() == headID {
					}
				}
			}
			return latch, true
		}
		return nil, false
	*/
}

func findOrigHead(iis []*Interval, head *Node) *Node {
	i, ok := getInterval(iis, head.DOTID())
	if !ok {
		return head
	}
	fmt.Println("#### [ interval lookup ] ####")
	fmt.Println("interval head:", head.DOTID())
	fmt.Println(i.String())
	fmt.Println("#### [/ interval lookup ] ####")
	return findOrigHead(iis, i.head)
}

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

// TODO: implement and use markNodesInLoop instead of loop. Get rid of n.inLoop
// in favour of loopHead?
//

// loop returns the nodes of the loop (latch, I.head), marking the loop header
// of each node.
//
// Pre: (latch, I.head) is a back-edge.
//
// Post: the nodes that belong to the loop (latch, I.head) are marked.
//
// ref: Figure 6-27; Cifuentes' Reverse Comilation Techniques.
func markNodesInLoop(g cfa.Graph, head, latch *Node, dom cfa.DominatorTree) []*Node {
	fmt.Println("=== [ markNodesInLoop] ===")
	nodesInLoop := []*Node{head}
	head.LoopHead = head
	for _, n := range ascRevPostOrder(NodesOf(g.Nodes())) {
		if head.RevPostNum < n.RevPostNum && n.RevPostNum <= latch.RevPostNum {
			fmt.Printf("dom %q: %v\n", n.DOTID(), dom.DominatorOf(n.ID()))
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
	fmt.Println("   [ nodes in loop ]")
	printNodes(nodesInLoop)
	fmt.Println("   [/ nodes in loop ]")
	fmt.Println()
	return nodesInLoop

	/*
		I := IIs[i][j]
		// The nodes belong to the same interval, since the interval header (i.e. x)
		// dominates all nodes of the interval, and in a loop, the loop header node
		// dominates all nodes of the loop. If a node belongs to a different
		// interval, it is not dominated by the loop header node, thus it cannot
		// belong to the same loop.

		//fmt.Println("head: ", I.head.DOTID()) // TODO: remove debug output
		//fmt.Println("latch:", latch.DOTID()) // TODO: remove debug output
		//fmt.Println() // TODO: remove debug output
		//fmt.Println(I.String()) // TODO: remove debug output
		// nodesInLoop := {x}
		nodesInLoop := []*Node{I.head}
		// loopHead(x) = x
		I.head.LoopHead = I.head
		// for (all nodes n \in {x + 1 ... y})
		for nodes := I.Nodes(); nodes.Next(); {
			// if n \in I(x)
			n := nodes.Node().(*Node)
			// The loop is formed of all nodes that are between x and y in terms of
			// node numbering.
			if I.head.RevPostNum < n.RevPostNum && n.RevPostNum <= latch.RevPostNum {
				// nodesInLoop = nodesInLoop \union {n}
				nodesInLoop = append(nodesInLoop, n)
				// if (loopHead(n) == No_Node)
				if n.LoopHead == nil {
					// loopHead(n) = x
					n.LoopHead = I.head
				}
			}
		}
		return nodesInLoop
	*/
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
	fmt.Println("=== [ nodes in loop ] ===")
	printNodes(nodesInLoop)
	fmt.Println()
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

// TODO: remove debug output.

func printNodes(ns []*Node) {
	for _, n := range ns {
		printNode(n)
	}
}

func printNode(n *Node) {
	fmt.Println("Node:      ", n.Node.DOTID())
	fmt.Println("PreNum:    ", n.PreNum)
	fmt.Println("RevPostNum:", n.RevPostNum)
	if n.LoopHead != nil {
		fmt.Println("LoopHead:  ", n.LoopHead.DOTID())
	}
	fmt.Println("LoopType:  ", n.LoopType)
	if n.LoopFollow != nil {
		fmt.Println("LoopFollow:", n.LoopFollow.DOTID())
	}
	fmt.Println()
}

// sortNodes sorts the list of nodes by DOTID.
func sortNodes(ns []*Node) []*Node {
	less := func(i, j int) bool {
		return natsort.Less(ns[i].DOTID(), ns[j].DOTID())
	}
	sort.Slice(ns, less)
	return ns
}
