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
func (fgen *funcGen) decompileFuncDef(irFunc *ir.Function) {
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
		fgen.liftBasicBlock(block.BasicBlock, liftTerm)
	case *IfElse:
		fgen.liftIfElse(block)
	default:
		panic(fmt.Errorf("support for pseudo basic block type %T not yet implemented", block))
	}
}

// liftBasicBlock lifts the LLVM IR basic block to Go source code, emitting to
// f. The liftTerm parameter determines whether to lift or skip the terminator
// instruction of the basic block.
func (fgen *funcGen) liftBasicBlock(block *ir.BasicBlock, liftTerm bool) {
	for _, inst := range block.Insts {
		fgen.liftInst(inst)
	}
	if liftTerm {
		fgen.liftTerm(block.Term)
	}
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
	fgen.liftBlock(block.Exit, true)
}

// primBlocks returns the list of pseudo basic blocks corresponding to the
// recovered high-level primitives of the given function.
func (fgen *funcGen) primBlocks(irFunc *ir.Function) []Block {
	prims, err := fgen.gen.Prims(irFunc)
	if err != nil {
		fgen.gen.eh(err)
		// Continue with recovery, even on error.
	}
	blocks := make(map[string]Block)
	for _, block := range irFunc.Blocks {
		blocks[block.Name()] = &IRBlock{BasicBlock: block}
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
			_ = entry
			_ = exit
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
	*ir.BasicBlock
}

func (block *IRBlock) GetTerm() ir.Terminator {
	return block.Term
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
// Unnamed
func newIdent(v namedValue) *ast.Ident {
	if v.IsUnnamed() {
		name := fmt.Sprintf("_%d", v.ID())
		return ast.NewIdent(name)
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
	return ast.NewIdent(name)
}
