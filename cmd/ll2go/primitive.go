// TODO: Verify that the if_return.dot primitive correctly maps
//    if cond {
//       return
//    }
// and not
//    if cond {
//       f()
//    }

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"

	"decomp.org/x/graphs"
	xprimitive "decomp.org/x/graphs/primitive"
	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
	"llvm.org/llvm/bindings/go/llvm"
)

// primitive represents a control flow primitive, such as a 2-way conditional, a
// pre-test loop or a single statement or a list of statements. Each primitive
// conceptually represents a basic block and may be treated as an instruction or
// a statement of other basic blocks.
type primitive struct {
	// The control flow primitive is conceptually a basic block, and as such
	// requires a basic block name.
	name string
	// Statements of the control flow primitive.
	stmts []ast.Stmt
	// Terminator instruction.
	term llvm.Value
}

// Name returns the name of the primitive, which conceptually represents a basic
// block.
func (prim *primitive) Name() string { return prim.name }

// Stmts returns the statements of the primitive, which conceptually represents
// a basic block.
func (prim *primitive) Stmts() []ast.Stmt { return prim.stmts }

// SetStmts sets the statements of the primitive, which conceptually represents
// a basic block.
func (prim *primitive) SetStmts(stmts []ast.Stmt) { prim.stmts = stmts }

// Term returns the terminator instruction of the primitive, which conceptually
// represents a basic block.
func (prim *primitive) Term() llvm.Value { return prim.term }

// restructure attempts to create a structured control flow for a function based
// on the provided control flow graph (which contains one node per basic block)
// and the function's basic blocks. It does so by repeatedly locating and
// merging structured subgraphs into single nodes until the entire graph is
// reduced into a single node or no structured subgraphs may be located.
func restructure(graph *dot.Graph, bbs map[string]BasicBlock, hprims []*xprimitive.Primitive) (*ast.BlockStmt, error) {
	for _, hprim := range hprims {
		subName := hprim.Prim // identified primitive; e.g. "if", "if_else"
		m := hprim.Nodes      // node mapping
		newName := hprim.Node // new node name

		// Create a control flow primitive based on the identified subgraph.
		primBBs := make(map[string]BasicBlock)
		for _, gname := range m {
			bb, ok := bbs[gname]
			if !ok {
				return nil, errutil.Newf("unable to locate basic block %q", gname)
			}
			primBBs[gname] = bb
			delete(bbs, gname)
		}
		prim, err := createPrim(subName, m, primBBs, newName)
		if err != nil {
			return nil, errutil.Err(err)
		}
		if flagVerbose && !flagQuiet {
			fmt.Println("located primitive:")
			printBB(prim)
		}
		bbs[prim.Name()] = prim
	}

	for _, bb := range bbs {
		if !bb.Term().IsNil() {
			// TODO: Remove debug output.
			bb.Term().Dump()
			return nil, errutil.Newf("invalid terminator instruction of last basic block in function; expected nil since return statements are already handled")
		}
		block := &ast.BlockStmt{
			List: bb.Stmts(),
		}
		return block, nil
	}
	return nil, errutil.New("unable to locate basic block")
}

// createPrim creates a control flow primitive based on the identified subgraph
// and its node pair mapping and basic blocks. The new control flow primitive
// conceptually forms a new basic block with the specified name.
func createPrim(subName string, m map[string]string, bbs map[string]BasicBlock, newName string) (*primitive, error) {
	switch subName {
	case "if":
		return createIfPrim(m, bbs, newName)
	case "if_else":
		return createIfElsePrim(m, bbs, newName)
	case "if_return":
		return createIfPrim(m, bbs, newName)
	case "list":
		return createListPrim(m, bbs, newName)
	case "post_loop":
		return createPostLoopPrim(m, bbs, newName)
	case "pre_loop":
		return createPreLoopPrim(m, bbs, newName)
	default:
		return nil, errutil.Newf("control flow primitive of subgraph %q not yet supported", subName)
	}
}

// createListPrim creates a list primitive containing a slice of Go statements
// based on the identified subgraph, its node pair mapping and its basic blocks.
// The new control flow primitive conceptually represents a basic block with the
// given name.
//
// Contents of "list.dot":
//
//    digraph list {
//       entry [label="entry"]
//       exit [label="exit"]
//       entry->exit
//    }
func createListPrim(m map[string]string, bbs map[string]BasicBlock, newName string) (*primitive, error) {
	// Locate graph nodes.
	nameEntry, ok := m["entry"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "entry"`)
	}
	nameExit, ok := m["exit"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "exit"`)
	}
	bbEntry, ok := bbs[nameEntry]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameEntry)
	}
	bbExit, ok := bbs[nameExit]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameExit)
	}

	// Create and return new primitive.
	//
	//    entry
	//    exit
	stmts := append(bbEntry.Stmts(), bbExit.Stmts()...)
	prim := &primitive{
		name:  newName,
		stmts: stmts,
		term:  bbExit.Term(),
	}
	return prim, nil
}

// createIfPrim creates an if-statement primitive based on the identified
// subgraph, its node pair mapping and its basic blocks. The new control flow
// primitive conceptually represents a basic block with the given name.
//
// Contents of "if.dot":
//
//    digraph if {
//       cond [label="entry"]
//       body
//       exit [label="exit"]
//       cond->body [label="true"]
//       cond->exit [label="false"]
//       body->exit
//    }
func createIfPrim(m map[string]string, bbs map[string]BasicBlock, newName string) (*primitive, error) {
	// Locate graph nodes.
	nameCond, ok := m["cond"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "cond"`)
	}
	nameBody, ok := m["body"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "body"`)
	}
	nameExit, ok := m["exit"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "exit"`)
	}
	bbCond, ok := bbs[nameCond]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameCond)
	}
	bbBody, ok := bbs[nameBody]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameBody)
	}
	bbExit, ok := bbs[nameExit]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameExit)
	}

	// Create and return new primitive.
	//
	//    cond_stmts
	//    if cond {
	//       body
	//    }
	//    exit

	// Create if-statement.
	cond, _, _, err := getBrCond(bbCond.Term())
	if err != nil {
		return nil, errutil.Err(err)
	}
	ifStmt := &ast.IfStmt{
		Cond: cond,
		Body: &ast.BlockStmt{List: bbBody.Stmts()},
	}

	// Create primitive.
	stmts := append(bbCond.Stmts(), ifStmt)
	stmts = append(stmts, bbExit.Stmts()...)
	prim := &primitive{
		name:  newName,
		stmts: stmts,
		term:  bbExit.Term(),
	}
	return prim, nil
}

// createIfElsePrim creates an if-else primitive based on the identified
// subgraph, its node pair mapping and its basic blocks. The new control flow
// primitive conceptually represents a basic block with the given name.
//
// Contents of "if_else.dot":
//
//    digraph if_else {
//       cond [label="entry"]
//       body_true
//       body_false
//       exit [label="exit"]
//       cond->body_true [label="true"]
//       cond->body_false [label="false"]
//       body_true->exit
//       body_false->exit
//    }
func createIfElsePrim(m map[string]string, bbs map[string]BasicBlock, newName string) (*primitive, error) {
	// Locate graph nodes.
	nameCond, ok := m["cond"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "cond"`)
	}
	nameBodyTrue, ok := m["body_true"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "body_true"`)
	}
	nameBodyFalse, ok := m["body_false"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "body_false"`)
	}
	nameExit, ok := m["exit"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "exit"`)
	}
	bbCond, ok := bbs[nameCond]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameCond)
	}

	// The body nodes (body_true and body_false) of if-else primitives are
	// indistinguishable at the graph level. Verify their names against the
	// terminator instruction of the basic block and swap them if necessary.
	cond, targetTrue, targetFalse, err := getBrCond(bbCond.Term())
	if err != nil {
		return nil, errutil.Err(err)
	}
	if targetTrue != nameBodyTrue && targetTrue != nameBodyFalse {
		return nil, errutil.Newf("invalid target true branch; got %q, expected %q or %q", targetTrue, nameBodyTrue, nameBodyFalse)
	}
	if targetFalse != nameBodyTrue && targetFalse != nameBodyFalse {
		return nil, errutil.Newf("invalid target false branch; got %q, expected %q or %q", targetFalse, nameBodyTrue, nameBodyFalse)
	}
	nameBodyTrue = targetTrue
	nameBodyFalse = targetFalse

	bbBodyTrue, ok := bbs[nameBodyTrue]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameBodyTrue)
	}
	bbBodyFalse, ok := bbs[nameBodyFalse]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameBodyFalse)
	}
	bbExit, ok := bbs[nameExit]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameExit)
	}

	// Create and return new primitive.
	//
	//    cond_stmts
	//    if cond {
	//       body_true
	//    } else {
	//       body_false
	//    }
	//    exit

	// Create if-else statement.
	ifElseStmt := &ast.IfStmt{
		Cond: cond,
		Body: &ast.BlockStmt{List: bbBodyTrue.Stmts()},
		Else: &ast.BlockStmt{List: bbBodyFalse.Stmts()},
	}

	// Create primitive.
	stmts := append(bbCond.Stmts(), ifElseStmt)
	stmts = append(stmts, bbExit.Stmts()...)
	prim := &primitive{
		name:  newName,
		stmts: stmts,
		term:  bbExit.Term(),
	}
	return prim, nil
}

// createPreLoopPrim creates a pre-test loop primitive based on the identified
// subgraph, its node pair mapping and its basic blocks. The new control flow
// primitive conceptually represents a basic block with the given name.
//
// Contents of "pre_loop.dot":
//
//    digraph pre_loop {
//       cond [label="entry"]
//       body
//       exit [label="exit"]
//       cond->body [label="true"]
//       body->cond
//       cond->exit [label="false"]
//    }
func createPreLoopPrim(m map[string]string, bbs map[string]BasicBlock, newName string) (*primitive, error) {
	// Locate graph nodes.
	nameCond, ok := m["cond"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "cond"`)
	}
	nameBody, ok := m["body"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "body"`)
	}
	nameExit, ok := m["exit"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "exit"`)
	}
	bbCond, ok := bbs[nameCond]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameCond)
	}
	bbBody, ok := bbs[nameBody]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameBody)
	}
	bbExit, ok := bbs[nameExit]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameExit)
	}

	// Locate and expand the condition.
	//
	//    // from:
	//    _2 := i < 10
	//    if _2 {
	//
	//    // to:
	//    if i < 10 {
	cond, _, _, err := getBrCond(bbCond.Term())
	if err != nil {
		return nil, errutil.Err(err)
	}
	cond, err = expand(bbCond, cond)
	if err != nil {
		return nil, errutil.Err(err)
	}

	if len(bbCond.Stmts()) != 0 {
		// Produce the following primitive instead of a regular for loop if cond
		// contains statements.
		//
		//    for {
		//       cond_stmts
		//       if !cond {
		//          break
		//       }
		//       body_true
		//    }
		//    exit

		// Create if-statement.
		ifStmt := &ast.IfStmt{
			Cond: &ast.UnaryExpr{Op: token.NOT, X: cond}, // negate condition
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
		}

		// Create for-loop.
		body := append(bbCond.Stmts(), ifStmt)
		body = append(body, bbBody.Stmts()...)
		forStmt := &ast.ForStmt{
			Body: &ast.BlockStmt{List: body},
		}

		// Create primitive.
		stmts := []ast.Stmt{forStmt}
		stmts = append(stmts, bbExit.Stmts()...)
		prim := &primitive{
			name:  newName,
			stmts: stmts,
			term:  bbExit.Term(),
		}
		return prim, nil
	}

	// Create and return new primitive.
	//
	//    for cond {
	//       body_true
	//    }
	//    exit

	// Create for-loop.
	forStmt := &ast.ForStmt{
		Cond: cond,
		Body: &ast.BlockStmt{List: bbBody.Stmts()},
	}

	// Create primitive.
	stmts := []ast.Stmt{forStmt}
	stmts = append(stmts, bbExit.Stmts()...)
	prim := &primitive{
		name:  newName,
		stmts: stmts,
		term:  bbExit.Term(),
	}
	return prim, nil
}

// createPostLoopPrim creates a post-test loop primitive based on the identified
// subgraph, its node pair mapping and its basic blocks. The new control flow
// primitive conceptually represents a basic block with the given name.
//
// Contents of "post_loop.dot":
//
//    digraph post_loop {
//       cond [label="entry"]
//       exit [label="exit"]
//       cond->cond [label="true"]
//       cond->exit [label="false"]
//    }
func createPostLoopPrim(m map[string]string, bbs map[string]BasicBlock, newName string) (*primitive, error) {
	// Locate graph nodes.
	nameCond, ok := m["cond"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "cond"`)
	}
	nameExit, ok := m["exit"]
	if !ok {
		return nil, errutil.New(`unable to locate node pair for sub node "exit"`)
	}
	bbCond, ok := bbs[nameCond]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameCond)
	}
	bbExit, ok := bbs[nameExit]
	if !ok {
		return nil, errutil.Newf("unable to locate basic block %q", nameExit)
	}

	// Create and return new primitive.
	//
	//    for {
	//       cond_stmts
	//       if !cond {
	//          break
	//       }
	//    }
	//    exit

	// Create if-statement.
	cond, _, _, err := getBrCond(bbCond.Term())
	if err != nil {
		return nil, errutil.Err(err)
	}
	ifStmt := &ast.IfStmt{
		Cond: &ast.UnaryExpr{Op: token.NOT, X: cond}, // negate condition
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
	}

	// Create for-loop.
	body := bbCond.Stmts()
	body = append(body, ifStmt)
	forStmt := &ast.ForStmt{
		Body: &ast.BlockStmt{List: body},
	}

	// Create primitive.
	stmts := []ast.Stmt{forStmt}
	stmts = append(stmts, bbExit.Stmts()...)
	prim := &primitive{
		name:  newName,
		stmts: stmts,
		term:  bbExit.Term(),
	}
	return prim, nil
}

// printMapping prints the mapping from sub node name to graph node name for an
// isomorphism of sub in graph.
func printMapping(graph *dot.Graph, sub *graphs.SubGraph, m map[string]string) {
	entry := m[sub.Entry()]
	var snames []string
	for sname := range m {
		snames = append(snames, sname)
	}
	sort.Strings(snames)
	fmt.Printf("Isomorphism of %q found at node %q:\n", sub.Name, entry)
	for _, sname := range snames {
		fmt.Printf("   %q=%q\n", sname, m[sname])
	}
}
