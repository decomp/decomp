package decompile

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/value"
)

// decompileFuncDef decompiles the LLVM IR function definition to Go source
// code, emitting to f.
func (fgen *funcGen) decompileFuncDef(irFunc *ir.Func) {
	blockStmt := &ast.BlockStmt{}
	fgen.f.Body = blockStmt
	fgen.cur = blockStmt
	blocks := fgen.primBlocks(irFunc)
	for _, block := range blocks {
		fgen.liftBlock(block, true)
	}
}

// liftBlock lifts the pseudo basic block to Go source code, emitting to f. The
// liftTerm parameter determines whether to lift or skip the terminator
// instruction of the basic block.
func (fgen *funcGen) liftBlock(block Block, liftTerm bool) {
	switch block := block.(type) {
	case *IRBlock:
		fgen.liftBasicBlock(block.Block, liftTerm)
	case *Seq:
		fgen.liftSeq(block)
	case *If:
		fgen.liftIf(block)
	case *IfElse:
		fgen.liftIfElse(block)
	case *PreLoop:
		fgen.liftPreLoop(block)
	default:
		panic(fmt.Errorf("support for pseudo basic block type %T not yet implemented", block))
	}
}

// liftBasicBlock lifts the LLVM IR basic block to Go source code, emitting to
// f. The liftTerm parameter determines whether to lift or skip the terminator
// instruction of the basic block.
func (fgen *funcGen) liftBasicBlock(block *ir.Block, liftTerm bool) {
	for _, inst := range block.Insts {
		fgen.liftInst(inst)
	}
	if liftTerm {
		fgen.liftTerm(block.Term)
	}
}

// liftSeq lifts the pseudo sequence block to Go source code, emitting to f.
func (fgen *funcGen) liftSeq(block *Seq) {
	// Lift entry block.
	fgen.liftBlock(block.Entry, false)
	// Lift exit block.
	fgen.liftBlock(block.Exit, false)
}

// liftIf lifts the pseudo if block to Go source code, emitting to f.
func (fgen *funcGen) liftIf(block *If) {
	// Lift cond block.
	fgen.liftBlock(block.Cond, false)
	// Get if-else statement.
	body := &ast.BlockStmt{}
	ifStmt := &ast.IfStmt{
		Cond: fgen.getCond(block.Cond.GetTerm()),
		Body: body,
	}
	fgen.cur.List = append(fgen.cur.List, ifStmt)
	cur := fgen.cur
	// Lift body block.
	fgen.cur = body
	fgen.liftBlock(block.Body, false)
	// Lift exit block.
	fgen.cur = cur
	fgen.liftBlock(block.Exit, false)
}

// liftIfElse lifts the pseudo if-else block to Go source code, emitting to f.
func (fgen *funcGen) liftIfElse(block *IfElse) {
	// Lift cond block.
	fgen.liftBlock(block.Cond, false)
	// Get if-else statement.
	bodyTrue := &ast.BlockStmt{}
	bodyFalse := &ast.BlockStmt{}
	ifStmt := &ast.IfStmt{
		Cond: fgen.getCond(block.Cond.GetTerm()),
		Body: bodyTrue,
		Else: bodyFalse,
	}
	fgen.cur.List = append(fgen.cur.List, ifStmt)
	cur := fgen.cur
	// Lift body true block.
	fgen.cur = bodyTrue
	fgen.liftBlock(block.BodyTrue, false)
	// Lift body false block.
	fgen.cur = bodyFalse
	fgen.liftBlock(block.BodyFalse, false)
	// Lift exit block.
	fgen.cur = cur
	fgen.liftBlock(block.Exit, false)
}

// liftPreLoop lifts the pseudo pre-loop block to Go source code, emitting to f.
func (fgen *funcGen) liftPreLoop(block *PreLoop) {
	// Lift cond block.
	fgen.liftBlock(block.Cond, false)
	// Get if-else statement.
	body := &ast.BlockStmt{}
	forStmt := &ast.ForStmt{
		Cond: fgen.getCond(block.Cond.GetTerm()),
		Body: body,
	}
	fgen.cur.List = append(fgen.cur.List, forStmt)
	cur := fgen.cur
	// Lift body block.
	fgen.cur = body
	fgen.liftBlock(block.Body, false)
	// Lift exit block.
	fgen.cur = cur
	fgen.liftBlock(block.Exit, false)
}

// primBlocks returns the list of pseudo basic blocks corresponding to the
// recovered high-level primitives of the given function.
func (fgen *funcGen) primBlocks(irFunc *ir.Func) []Block {
	prims, err := fgen.gen.Prims(irFunc)
	if err != nil {
		fgen.gen.eh(err)
		// Continue with recovery, even on error.
	}
	blocks := make(map[string]Block)
	for _, block := range irFunc.Blocks {
		blocks[block.Name()] = &IRBlock{Block: block}
	}
	for _, prim := range prims {
		dbg.Printf("recovering %q primitive", prim.Prim)
		switch prim.Prim {
		case "seq":
			entryName := prim.Nodes["entry"]
			entry, ok := blocks[entryName]
			if !ok {
				fgen.gen.Errorf("unable to locate entry block %q of primitive %q in function %q", entryName, prim.Prim, irFunc.Name())
				continue
			}
			exitName := prim.Nodes["exit"]
			exit, ok := blocks[exitName]
			if !ok {
				fgen.gen.Errorf("unable to locate exit block %q of primitive %q in function %q", exitName, prim.Prim, irFunc.Name())
				continue
			}
			block := &Seq{
				BlockName: prim.Entry,
				Entry:     entry,
				Exit:      exit,
			}
			delete(blocks, entryName)
			delete(blocks, exitName)
			blocks[block.Name()] = block
		case "if":
			condName := prim.Nodes["cond"]
			cond, ok := blocks[condName]
			if !ok {
				fgen.gen.Errorf("unable to locate cond block %q of primitive %q in function %q", condName, prim.Prim, irFunc.Name())
				continue
			}
			bodyName := prim.Nodes["body"]
			body, ok := blocks[bodyName]
			if !ok {
				fgen.gen.Errorf("unable to locate body block %q of primitive %q in function %q", bodyName, prim.Prim, irFunc.Name())
				continue
			}
			exitName := prim.Nodes["exit"]
			exit, ok := blocks[exitName]
			if !ok {
				fgen.gen.Errorf("unable to locate exit block %q of primitive %q in function %q", exitName, prim.Prim, irFunc.Name())
				continue
			}
			block := &If{
				BlockName: prim.Entry,
				Cond:      cond,
				Body:      body,
				Exit:      exit,
			}
			delete(blocks, condName)
			delete(blocks, bodyName)
			delete(blocks, exitName)
			blocks[block.Name()] = block
		case "if_else":
			condName := prim.Nodes["cond"]
			cond, ok := blocks[condName]
			if !ok {
				fgen.gen.Errorf("unable to locate cond block %q of primitive %q in function %q", condName, prim.Prim, irFunc.Name())
				continue
			}
			bodyTrueName := prim.Nodes["body_true"]
			bodyTrue, ok := blocks[bodyTrueName]
			if !ok {
				fgen.gen.Errorf("unable to locate body_true block %q of primitive %q in function %q", bodyTrueName, prim.Prim, irFunc.Name())
				continue
			}
			bodyFalseName := prim.Nodes["body_false"]
			bodyFalse, ok := blocks[bodyFalseName]
			if !ok {
				fgen.gen.Errorf("unable to locate body_false block %q of primitive %q in function %q", bodyFalseName, prim.Prim, irFunc.Name())
				continue
			}
			exitName := prim.Nodes["exit"]
			exit, ok := blocks[exitName]
			if !ok {
				fgen.gen.Errorf("unable to locate exit block %q of primitive %q in function %q", exitName, prim.Prim, irFunc.Name())
				continue
			}
			block := &IfElse{
				BlockName: prim.Entry,
				Cond:      cond,
				BodyTrue:  bodyTrue,
				BodyFalse: bodyFalse,
				Exit:      exit,
			}
			delete(blocks, condName)
			delete(blocks, bodyTrueName)
			delete(blocks, bodyFalseName)
			delete(blocks, exitName)
			blocks[block.Name()] = block
		case "pre_loop":
			condName := prim.Nodes["cond"]
			cond, ok := blocks[condName]
			if !ok {
				fgen.gen.Errorf("unable to locate cond block %q of primitive %q in function %q", condName, prim.Prim, irFunc.Name())
				continue
			}
			bodyName := prim.Nodes["body"]
			body, ok := blocks[bodyName]
			if !ok {
				fgen.gen.Errorf("unable to locate body block %q of primitive %q in function %q", bodyName, prim.Prim, irFunc.Name())
				continue
			}
			exitName := prim.Nodes["exit"]
			exit, ok := blocks[exitName]
			if !ok {
				fgen.gen.Errorf("unable to locate exit block %q of primitive %q in function %q", exitName, prim.Prim, irFunc.Name())
				continue
			}
			block := &PreLoop{
				BlockName: prim.Entry,
				Cond:      cond,
				Body:      body,
				Exit:      exit,
			}
			delete(blocks, condName)
			delete(blocks, bodyName)
			delete(blocks, exitName)
			blocks[block.Name()] = block
		default:
			panic(fmt.Errorf("support for primitive %q not yet implemented", prim.Prim))
		}
	}
	// Convert blocks to linear representation.
	var bbs []Block
	for _, bb := range irFunc.Blocks {
		if block, ok := blocks[bb.Name()]; ok {
			bbs = append(bbs, block)
		}
	}
	if len(blocks) != len(bbs) {
		panic(fmt.Errorf("number of recovered blocks mismatch; expected %d, got %d", len(blocks), len(bbs)))
	}
	return bbs
}

type Block interface {
	Name() string
	GetTerm() ir.Terminator
}

type IRBlock struct {
	*ir.Block
}

func (block *IRBlock) GetTerm() ir.Terminator {
	return block.Term
}

type PreLoop struct {
	BlockName string
	Cond      Block
	Body      Block
	Exit      Block
}

func (block *PreLoop) Name() string {
	return block.BlockName
}

func (block *PreLoop) GetTerm() ir.Terminator {
	return block.Exit.GetTerm()
}

type Seq struct {
	BlockName string
	Entry     Block
	Exit      Block
}

func (block *Seq) Name() string {
	return block.BlockName
}

func (block *Seq) GetTerm() ir.Terminator {
	return block.Exit.GetTerm()
}

type If struct {
	BlockName string
	Cond      Block
	Body      Block
	Exit      Block
}

func (block *If) Name() string {
	return block.BlockName
}

func (block *If) GetTerm() ir.Terminator {
	return block.Exit.GetTerm()
}

type IfElse struct {
	BlockName string
	Cond      Block
	BodyTrue  Block
	BodyFalse Block
	Exit      Block
}

func (block *IfElse) Name() string {
	return block.BlockName
}

func (block *IfElse) GetTerm() ir.Terminator {
	return block.Exit.GetTerm()
}

// liftInst lifts the LLVM IR instruction to Go source code, emitting to f.
func (fgen *funcGen) liftInst(inst ir.Instruction) {
	switch inst := inst.(type) {
	case *ir.InstAlloca:
		fgen.liftInstAlloca(inst)
	case *ir.InstStore:
		fgen.liftInstStore(inst)
	case *ir.InstLoad:
		fgen.liftInstLoad(inst)
	case *ir.InstICmp:
		fgen.liftInstICmp(inst)
	case *ir.InstAdd:
		// Variable name.
		name := newIdent(inst)
		// X and Y operands.
		x := fgen.liftValue(inst.X)
		y := fgen.liftValue(inst.Y)
		fgen.emitAssignBinOp(name, x, y, token.ADD)
	case *ir.InstMul:
		// Variable name.
		name := newIdent(inst)
		// X and Y operands.
		x := fgen.liftValue(inst.X)
		y := fgen.liftValue(inst.Y)
		fgen.emitAssignBinOp(name, x, y, token.MUL)
	default:
		panic(fmt.Errorf("support for instruction type %T not yet implemented", inst))
	}
}

// liftInstAlloca lifts the LLVM IR alloca instruction to Go source code,
// emitting to f.
func (fgen *funcGen) liftInstAlloca(inst *ir.InstAlloca) {
	// Variable name.
	name := newIdent(inst)
	// Element type.
	elemType, err := fgen.gen.goType(inst.ElemType)
	if err != nil {
		fgen.gen.eh(err)
		return
	}
	// (optional) Number of elements.
	var callExpr *ast.CallExpr
	if inst.NElems != nil {
		// Make slice of given length.
		nelems := fgen.liftValue(inst.NElems)
		sliceType := gotypes.NewSlice(elemType)
		callExpr = &ast.CallExpr{
			Fun: ast.NewIdent("make"),
			Args: []ast.Expr{
				goTypeExpr(sliceType),
				nelems,
			},
		}
	} else {
		// Allocate type using new.
		callExpr = &ast.CallExpr{
			Fun: ast.NewIdent("new"),
			Args: []ast.Expr{
				goTypeExpr(elemType),
			},
		}
	}
	// Append assignment statement.
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{name},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{callExpr},
	}
	fgen.cur.List = append(fgen.cur.List, assignStmt)
}

// liftInstStore lifts the LLVM IR store instruction to Go source code, emitting
// to f.
func (fgen *funcGen) liftInstStore(inst *ir.InstStore) {
	// Destination.
	dst := fgen.liftValue(inst.Dst)
	// Source.
	src := fgen.liftValue(inst.Src)
	// Append assignment statement.
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.StarExpr{X: dst}},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{src},
	}
	fgen.cur.List = append(fgen.cur.List, assignStmt)
}

// liftInstLoad lifts the LLVM IR load instruction to Go source code, emitting
// to f.
func (fgen *funcGen) liftInstLoad(inst *ir.InstLoad) {
	// Variable name.
	name := newIdent(inst)
	// Source.
	src := fgen.liftValue(inst.Src)
	// Append assignment statement.
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{name},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{&ast.StarExpr{X: src}},
	}
	fgen.cur.List = append(fgen.cur.List, assignStmt)
}

// liftInstICmp lifts the LLVM IR icmp instruction to Go source code, emitting
// to f.
func (fgen *funcGen) liftInstICmp(inst *ir.InstICmp) {
	// Variable name.
	name := newIdent(inst)
	// Predicate.
	op := ipred(inst.Pred)
	// X and Y operands.
	x := fgen.liftValue(inst.X)
	y := fgen.liftValue(inst.Y)
	binExpr := &ast.BinaryExpr{
		X:  x,
		Op: op,
		Y:  y,
	}
	// Append assignment statement.
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{name},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{binExpr},
	}
	fgen.cur.List = append(fgen.cur.List, assignStmt)
}

// emitAssignBinOp emits an assignment statement based on the given name,
// operands and binary operation, emitting to f.
func (fgen *funcGen) emitAssignBinOp(name *ast.Ident, x, y ast.Expr, op token.Token) {
	// Binary expression.
	binExpr := &ast.BinaryExpr{
		X:  x,
		Op: op,
		Y:  y,
	}
	// Append assignment statement.
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{name},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{binExpr},
	}
	fgen.cur.List = append(fgen.cur.List, assignStmt)
}

// ipred returns the Go token corresponding to the given LLVM IR integer
// comparison predicate.
func ipred(pred enum.IPred) token.Token {
	// TODO: figure out how to distinguish signed vs. unsigned values.
	switch pred {
	case enum.IPredEQ:
		return token.EQL
	case enum.IPredNE:
		return token.NEQ
	case enum.IPredSGE:
		return token.GEQ
	case enum.IPredSGT:
		return token.GTR
	case enum.IPredSLE:
		return token.LEQ
	case enum.IPredSLT:
		return token.LSS
	case enum.IPredUGE:
		return token.GEQ
	case enum.IPredUGT:
		return token.GTR
	case enum.IPredULE:
		return token.LEQ
	case enum.IPredULT:
		return token.LSS
	default:
		panic(fmt.Errorf("support for integer comparison predicate %v not yet implemented", pred))
	}
}

// liftTerm lifts the LLVM IR terminator to Go source code, emitting to f.
func (fgen *funcGen) liftTerm(term ir.Terminator) {
	switch term := term.(type) {
	case *ir.TermRet:
		fgen.liftTermRet(term)
	case *ir.TermBr:
		fgen.liftTermBr(term)
	case *ir.TermCondBr:
		fgen.liftTermCondBr(term)
	default:
		panic(fmt.Errorf("support for terminator %T not yet implemented", term))
	}
}

// liftTermRet lifts the LLVM IR ret terminator to Go source code, emitting to
// f.
func (fgen *funcGen) liftTermRet(term *ir.TermRet) {
	var results []ast.Expr
	if term.X != nil {
		result := fgen.liftValue(term.X)
		results = append(results, result)
	}
	returnStmt := &ast.ReturnStmt{
		Results: results,
	}
	fgen.cur.List = append(fgen.cur.List, returnStmt)
}

// liftTermBr lifts the LLVM IR br terminator to Go source code, emitting to f.
func (fgen *funcGen) liftTermBr(term *ir.TermBr) {
	gotoStmt := &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: newIdent(term.Target),
	}
	fgen.cur.List = append(fgen.cur.List, gotoStmt)
}

// liftTermCondBr lifts the LLVM IR conditional br terminator to Go source code,
// emitting to f.
func (fgen *funcGen) liftTermCondBr(term *ir.TermCondBr) {
	// Get if-else statement.
	bodyTrue := &ast.BlockStmt{}
	bodyFalse := &ast.BlockStmt{}
	ifStmt := &ast.IfStmt{
		Cond: fgen.getCond(term),
		Body: bodyTrue,
		Else: bodyFalse,
	}
	fgen.cur.List = append(fgen.cur.List, ifStmt)
	// Lift body true block.
	gotoTrueStmt := &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: newIdent(term.TargetTrue),
	}
	bodyTrue.List = append(bodyTrue.List, gotoTrueStmt)
	// Lift body false block.
	gotoFalseStmt := &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: newIdent(term.TargetFalse),
	}
	bodyFalse.List = append(bodyFalse.List, gotoFalseStmt)
}

// getCond returns the Go condition expression used in conditional branching of
// the given LLVM IR terminator, emitting to f.
func (fgen *funcGen) getCond(term ir.Terminator) ast.Expr {
	switch term := term.(type) {
	case *ir.TermCondBr:
		return fgen.liftValue(term.Cond)
	default:
		panic(fmt.Errorf("support for terminator %T not yet implemented", term))
	}
}

// liftValue lifts the LLVM IR value to a corresponding Go expression, emitting
// to f.
func (fgen *funcGen) liftValue(v value.Value) ast.Expr {
	switch v := v.(type) {
	case namedValue:
		return newIdent(v)
	case *constant.Int:
		return &ast.BasicLit{Kind: token.INT, Value: v.X.String()}
	default:
		panic(fmt.Errorf("support for value %T not yet implemented", v))
	}
}

// namedValue is a global or local variable.
type namedValue interface {
	value.Named
	IsUnnamed() bool
	ID() int64
}

// newIdent returns a new Go identifier based on the given LLVM IR identifier.
func newIdent(v namedValue) *ast.Ident {
	return ast.NewIdent(newName(v))
}

// newName returns a new Go name based on the given LLVM IR identifier.
func newName(v namedValue) string {
	if v.IsUnnamed() {
		name := fmt.Sprintf("_%d", v.ID())
		return name
	}
	f := func(r rune) rune {
		const (
			lower = "abcdefghijklmnopqrstuvwxyz"
			upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			digit = "0123456789"
			valid = lower + upper + digit + "_"
		)
		if !strings.ContainsRune(valid, r) {
			return '_'
		}
		return r
	}
	name := strings.Map(f, v.Name())
	return name
}
