package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	irtypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// insts converts the given LLVM IR instructions to a corresponding list of Go
// statements.
func (d *decompiler) insts(insts []ir.Instruction) []ast.Stmt {
	var stmts []ast.Stmt
	for _, inst := range insts {
		switch inst := inst.(type) {
		case *ir.InstPhi:
			// PHI instructions are handled during the pre-processing of basic
			// blocks.
			continue
		case *ir.InstSelect:
			// A select instruction corresponds to more than one Go statement, thus
			// it is handled outside of d.inst.
			stmts = append(stmts, d.instSelect(inst)...)
			continue
		}
		stmts = append(stmts, d.inst(inst))
	}
	return stmts
}

// inst converts the given LLVM IR instruction to a corresponding Go statement.
func (d *decompiler) inst(inst ir.Instruction) ast.Stmt {
	switch inst := inst.(type) {
	// Binary instructions
	case *ir.InstAdd:
		return d.instAdd(inst)
	case *ir.InstFAdd:
		return d.instFAdd(inst)
	case *ir.InstSub:
		return d.instSub(inst)
	case *ir.InstFSub:
		return d.instFSub(inst)
	case *ir.InstMul:
		return d.instMul(inst)
	case *ir.InstFMul:
		return d.instFMul(inst)
	case *ir.InstUDiv:
		return d.instUDiv(inst)
	case *ir.InstSDiv:
		return d.instSDiv(inst)
	case *ir.InstFDiv:
		return d.instFDiv(inst)
	case *ir.InstURem:
		return d.instURem(inst)
	case *ir.InstSRem:
		return d.instSRem(inst)
	case *ir.InstFRem:
		return d.instFRem(inst)
	// Bitwise instructions
	case *ir.InstShl:
		return d.instShl(inst)
	case *ir.InstLShr:
		return d.instLShr(inst)
	case *ir.InstAShr:
		return d.instAShr(inst)
	case *ir.InstAnd:
		return d.instAnd(inst)
	case *ir.InstOr:
		return d.instOr(inst)
	case *ir.InstXor:
		return d.instXor(inst)
	// Vector instructions
	case *ir.InstExtractElement:
		return d.instExtractElement(inst)
	case *ir.InstInsertElement:
		return d.instInsertElement(inst)
	case *ir.InstShuffleVector:
		return d.instShuffleVector(inst)
	// Aggregate instructions
	case *ir.InstExtractValue:
		return d.instExtractValue(inst)
	case *ir.InstInsertValue:
		return d.instInsertValue(inst)
	// Memory instructions
	case *ir.InstAlloca:
		return d.instAlloca(inst)
	case *ir.InstLoad:
		return d.instLoad(inst)
	case *ir.InstStore:
		return d.instStore(inst)
	case *ir.InstGetElementPtr:
		return d.instGetElementPtr(inst)
	// Conversion instructions
	case *ir.InstTrunc:
		return d.instTrunc(inst)
	case *ir.InstZExt:
		return d.instZExt(inst)
	case *ir.InstSExt:
		return d.instSExt(inst)
	case *ir.InstFPTrunc:
		return d.instFPTrunc(inst)
	case *ir.InstFPExt:
		return d.instFPExt(inst)
	case *ir.InstFPToUI:
		return d.instFPToUI(inst)
	case *ir.InstFPToSI:
		return d.instFPToSI(inst)
	case *ir.InstUIToFP:
		return d.instUIToFP(inst)
	case *ir.InstSIToFP:
		return d.instSIToFP(inst)
	case *ir.InstPtrToInt:
		return d.instPtrToInt(inst)
	case *ir.InstIntToPtr:
		return d.instIntToPtr(inst)
	case *ir.InstBitCast:
		return d.instBitCast(inst)
	case *ir.InstAddrSpaceCast:
		return d.instAddrSpaceCast(inst)
	// Other instructions
	case *ir.InstICmp:
		return d.instICmp(inst)
	case *ir.InstFCmp:
		return d.instFCmp(inst)
	case *ir.InstPhi:
		// PHI instructions are handled by d.funcDecl.
		panic(fmt.Sprintf("unexpected phi instruction `%v`", inst))
	case *ir.InstSelect:
		// select instructions are handled by d.insts.
		panic(fmt.Sprintf("unexpected select instruction `%v`", inst))
	case *ir.InstCall:
		return d.instCall(inst)
	default:
		panic(fmt.Sprintf("support for instruction %T not yet implemented", inst))
	}
}

// instAdd converts the given LLVM IR add instruction to a corresponding Go
// statement.
func (d *decompiler) instAdd(inst *ir.InstAdd) ast.Stmt {
	expr := d.binaryOp(inst.X, token.ADD, inst.Y)
	return d.assign(inst.Name, expr)
}

// instFAdd converts the given LLVM IR fadd instruction to a corresponding Go
// statement.
func (d *decompiler) instFAdd(inst *ir.InstFAdd) ast.Stmt {
	expr := d.binaryOp(inst.X, token.ADD, inst.Y)
	return d.assign(inst.Name, expr)
}

// instSub converts the given LLVM IR sub instruction to a corresponding Go
// statement.
func (d *decompiler) instSub(inst *ir.InstSub) ast.Stmt {
	expr := d.binaryOp(inst.X, token.SUB, inst.Y)
	return d.assign(inst.Name, expr)
}

// instFSub converts the given LLVM IR fsub instruction to a corresponding Go
// statement.
func (d *decompiler) instFSub(inst *ir.InstFSub) ast.Stmt {
	expr := d.binaryOp(inst.X, token.SUB, inst.Y)
	return d.assign(inst.Name, expr)
}

// instMul converts the given LLVM IR mul instruction to a corresponding Go
// statement.
func (d *decompiler) instMul(inst *ir.InstMul) ast.Stmt {
	expr := d.binaryOp(inst.X, token.MUL, inst.Y)
	return d.assign(inst.Name, expr)

}

// instFMul converts the given LLVM IR fmul instruction to a corresponding Go
// statement.
func (d *decompiler) instFMul(inst *ir.InstFMul) ast.Stmt {
	expr := d.binaryOp(inst.X, token.MUL, inst.Y)
	return d.assign(inst.Name, expr)
}

// instUDiv converts the given LLVM IR udiv instruction to a corresponding Go
// statement.
func (d *decompiler) instUDiv(inst *ir.InstUDiv) ast.Stmt {
	expr := d.binaryOp(inst.X, token.QUO, inst.Y)
	return d.assign(inst.Name, expr)
}

// instSDiv converts the given LLVM IR sdiv instruction to a corresponding Go
// statement.
func (d *decompiler) instSDiv(inst *ir.InstSDiv) ast.Stmt {
	expr := d.binaryOp(inst.X, token.QUO, inst.Y)
	return d.assign(inst.Name, expr)
}

// instFDiv converts the given LLVM IR fdiv instruction to a corresponding Go
// statement.
func (d *decompiler) instFDiv(inst *ir.InstFDiv) ast.Stmt {
	expr := d.binaryOp(inst.X, token.QUO, inst.Y)
	return d.assign(inst.Name, expr)
}

// instURem converts the given LLVM IR urem instruction to a corresponding Go
// statement.
func (d *decompiler) instURem(inst *ir.InstURem) ast.Stmt {
	expr := d.binaryOp(inst.X, token.REM, inst.Y)
	return d.assign(inst.Name, expr)
}

// instSRem converts the given LLVM IR srem instruction to a corresponding Go
// statement.
func (d *decompiler) instSRem(inst *ir.InstSRem) ast.Stmt {
	expr := d.binaryOp(inst.X, token.REM, inst.Y)
	return d.assign(inst.Name, expr)
}

// instFRem converts the given LLVM IR frem instruction to a corresponding Go
// statement.
func (d *decompiler) instFRem(inst *ir.InstFRem) ast.Stmt {
	expr := d.binaryOp(inst.X, token.REM, inst.Y)
	return d.assign(inst.Name, expr)
}

// instShl converts the given LLVM IR shl instruction to a corresponding Go
// statement.
func (d *decompiler) instShl(inst *ir.InstShl) ast.Stmt {
	expr := d.binaryOp(inst.X, token.SHL, inst.Y)
	return d.assign(inst.Name, expr)
}

// instLShr converts the given LLVM IR lshr instruction to a corresponding Go
// statement.
func (d *decompiler) instLShr(inst *ir.InstLShr) ast.Stmt {
	expr := d.binaryOp(inst.X, token.SHR, inst.Y)
	return d.assign(inst.Name, expr)
}

// instAShr converts the given LLVM IR ashr instruction to a corresponding Go
// statement.
func (d *decompiler) instAShr(inst *ir.InstAShr) ast.Stmt {
	// TODO: Differentiate between logical shift right and arithmetic shift
	// right.
	expr := d.binaryOp(inst.X, token.SHR, inst.Y)
	return d.assign(inst.Name, expr)
}

// instAnd converts the given LLVM IR and instruction to a corresponding Go
// statement.
func (d *decompiler) instAnd(inst *ir.InstAnd) ast.Stmt {
	expr := d.binaryOp(inst.X, token.AND, inst.Y)
	return d.assign(inst.Name, expr)
}

// instOr converts the given LLVM IR or instruction to a corresponding Go
// statement.
func (d *decompiler) instOr(inst *ir.InstOr) ast.Stmt {
	expr := d.binaryOp(inst.X, token.OR, inst.Y)
	return d.assign(inst.Name, expr)
}

// instXor converts the given LLVM IR xor instruction to a corresponding Go
// statement.
func (d *decompiler) instXor(inst *ir.InstXor) ast.Stmt {
	expr := d.binaryOp(inst.X, token.XOR, inst.Y)
	return d.assign(inst.Name, expr)
}

// instExtractElement converts the given LLVM IR extractelement instruction to a
// corresponding Go statement.
func (d *decompiler) instExtractElement(inst *ir.InstExtractElement) ast.Stmt {
	src := &ast.IndexExpr{
		X:     d.value(inst.X),
		Index: d.value(inst.Index),
	}
	return d.assign(inst.Name, src)
}

// instInsertElement converts the given LLVM IR insertelement instruction to a
// corresponding Go statement.
func (d *decompiler) instInsertElement(inst *ir.InstInsertElement) ast.Stmt {
	// TODO: Implement insertelement.
	panic("not yet implemented")
}

// instShuffleVector converts the given LLVM IR shufflevector instruction to a
// corresponding Go statement.
func (d *decompiler) instShuffleVector(inst *ir.InstShuffleVector) ast.Stmt {
	// TODO: Implement shufflevector.
	panic("not yet implemented")
}

// instExtractValue converts the given LLVM IR extractvalue instruction to a
// corresponding Go statement.
func (d *decompiler) instExtractValue(inst *ir.InstExtractValue) ast.Stmt {
	src := d.value(inst.X)
	for _, index := range inst.Indices {
		src = &ast.IndexExpr{
			X:     src,
			Index: d.intLit(index),
		}
	}
	return d.assign(inst.Name, src)
}

// instInsertValue converts the given LLVM IR insertvalue instruction to a
// corresponding Go statement.
func (d *decompiler) instInsertValue(inst *ir.InstInsertValue) ast.Stmt {
	// TODO: Implement insertvalue.
	panic("not yet implemented")
}

// instAlloca converts the given LLVM IR alloca instruction to a corresponding
// Go statement.
func (d *decompiler) instAlloca(inst *ir.InstAlloca) ast.Stmt {
	typ := d.goType(inst.Elem)
	if inst.NElems != nil {
		typ = &ast.ArrayType{
			Len: d.value(inst.NElems),
			Elt: typ,
		}
	}
	expr := &ast.CallExpr{
		Fun:  ast.NewIdent("new"),
		Args: []ast.Expr{typ},
	}
	return d.assign(inst.Name, expr)
}

// instLoad converts the given LLVM IR load instruction to a corresponding Go
// statement.
func (d *decompiler) instLoad(inst *ir.InstLoad) ast.Stmt {
	// TODO: Handle type (inst.Typ).
	expr := &ast.StarExpr{
		X: d.value(inst.Src),
	}
	return d.assign(inst.Name, expr)
}

// instStore converts the given LLVM IR store instruction to a corresponding Go
// statement.
func (d *decompiler) instStore(inst *ir.InstStore) ast.Stmt {
	dst := &ast.StarExpr{
		X: d.value(inst.Dst),
	}
	return &ast.AssignStmt{
		Lhs: []ast.Expr{dst},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{d.value(inst.Src)},
	}
}

// instGetElementPtr converts the given LLVM IR getelementptr instruction to a
// corresponding Go statement.
func (d *decompiler) instGetElementPtr(inst *ir.InstGetElementPtr) ast.Stmt {
	src := d.value(inst.Src)
	// TODO: Validate if index expressions should be added in reverse order.
	for _, index := range inst.Indices {
		src = &ast.IndexExpr{
			X:     src,
			Index: d.value(index),
		}
	}
	expr := &ast.UnaryExpr{
		Op: token.AND,
		X:  src,
	}
	return d.assign(inst.Name, expr)
}

// instTrunc converts the given LLVM IR trunc instruction to a corresponding Go
// statement.
func (d *decompiler) instTrunc(inst *ir.InstTrunc) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instZExt converts the given LLVM IR zext instruction to a corresponding Go
// statement.
func (d *decompiler) instZExt(inst *ir.InstZExt) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instSExt converts the given LLVM IR sext instruction to a corresponding Go
// statement.
func (d *decompiler) instSExt(inst *ir.InstSExt) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instFPTrunc converts the given LLVM IR fptrunc instruction to a corresponding
// Go statement.
func (d *decompiler) instFPTrunc(inst *ir.InstFPTrunc) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instFPExt converts the given LLVM IR fpext instruction to a corresponding Go
// statement.
func (d *decompiler) instFPExt(inst *ir.InstFPExt) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instFPToUI converts the given LLVM IR fptoui instruction to a corresponding
// Go statement.
func (d *decompiler) instFPToUI(inst *ir.InstFPToUI) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instFPToSI converts the given LLVM IR fptosi instruction to a corresponding
// Go statement.
func (d *decompiler) instFPToSI(inst *ir.InstFPToSI) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instUIToFP converts the given LLVM IR uitofp instruction to a corresponding
// Go statement.
func (d *decompiler) instUIToFP(inst *ir.InstUIToFP) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instSIToFP converts the given LLVM IR sitofp instruction to a corresponding
// Go statement.
func (d *decompiler) instSIToFP(inst *ir.InstSIToFP) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instPtrToInt converts the given LLVM IR ptrtoint instruction to a
// corresponding Go statement.
func (d *decompiler) instPtrToInt(inst *ir.InstPtrToInt) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instIntToPtr converts the given LLVM IR inttoptr instruction to a
// corresponding Go statement.
func (d *decompiler) instIntToPtr(inst *ir.InstIntToPtr) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instBitCast converts the given LLVM IR bitcast instruction to a corresponding
// Go statement.
func (d *decompiler) instBitCast(inst *ir.InstBitCast) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instAddrSpaceCast converts the given LLVM IR addrspacecast instruction to a
// corresponding Go statement.
func (d *decompiler) instAddrSpaceCast(inst *ir.InstAddrSpaceCast) ast.Stmt {
	expr := d.convert(inst.From, inst.To)
	return d.assign(inst.Name, expr)
}

// instICmp converts the given LLVM IR icmp instruction to a corresponding Go
// statement.
func (d *decompiler) instICmp(inst *ir.InstICmp) ast.Stmt {
	op := intPred(inst.Pred)
	expr := d.binaryOp(inst.X, op, inst.Y)
	return d.assign(inst.Name, expr)
}

// instFCmp converts the given LLVM IR fcmp instruction to a corresponding Go
// statement.
func (d *decompiler) instFCmp(inst *ir.InstFCmp) ast.Stmt {
	op := floatPred(inst.Pred)
	expr := d.binaryOp(inst.X, op, inst.Y)
	return d.assign(inst.Name, expr)
}

// instSelect converts the given LLVM IR select instruction to a corresponding
// Go statement.
func (d *decompiler) instSelect(inst *ir.InstSelect) []ast.Stmt {
	spec := &ast.ValueSpec{
		Names: []*ast.Ident{d.localIdent(inst.Name)},
		Type:  d.goType(inst.X.Type()),
	}
	declStmt := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{spec},
		},
	}
	ifStmt := &ast.IfStmt{
		Cond: d.value(inst.Cond),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{d.assign(inst.Name, d.value(inst.X))},
		},
		Else: &ast.BlockStmt{
			List: []ast.Stmt{d.assign(inst.Name, d.value(inst.Y))},
		},
	}
	return []ast.Stmt{declStmt, ifStmt}
}

// instCall converts the given LLVM IR call instruction to a corresponding Go
// statement.
func (d *decompiler) instCall(inst *ir.InstCall) ast.Stmt {
	var callee ast.Expr
	switch c := inst.Callee.(type) {
	case *ir.Function:
		// global function identifier.
		callee = d.globalIdent(c.Name)
	case *irtypes.Param:
		// local function identifier.
		callee = d.localIdent(c.Name)
	case *constant.ExprBitCast:
		callee = d.value(c)
	case *ir.InstBitCast:
		callee = d.value(c)
	case *ir.InstLoad:
		callee = d.value(c)
	default:
		panic(fmt.Sprintf("support for callee type %T not yet implemented", c))
	}
	var args []ast.Expr
	for _, a := range inst.Args {
		args = append(args, d.value(a))
	}
	expr := &ast.CallExpr{
		Fun:  callee,
		Args: args,
	}
	if irtypes.Equal(inst.Sig.Ret, irtypes.Void) {
		return &ast.ExprStmt{X: expr}
	}
	return d.assign(inst.Name, expr)
}

// binaryOp converts the given LLVM IR binary operation to a corresponding Go
// expression.
func (d *decompiler) binaryOp(x value.Value, op token.Token, y value.Value) ast.Expr {
	return &ast.BinaryExpr{
		X:  d.value(x),
		Op: op,
		Y:  d.value(y),
	}
}

// convert returns a Go statement for converting the given LLVM IR value into
// the specified type.
func (d *decompiler) convert(from value.Value, to irtypes.Type) ast.Expr {
	// Type conversion represented as a Go call expression.
	return &ast.CallExpr{
		Fun:  d.goType(to),
		Args: []ast.Expr{d.value(from)},
	}
}

// assign returns an assignment statement, assigning expr to the given local
// variable.
func (d *decompiler) assign(name string, expr ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{d.localIdent(name)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{expr},
	}
}

// intPred converts the given LLVM IR integer predicate to a corresponding Go
// token.
func intPred(pred ir.IntPred) token.Token {
	// TODO: Differentiate between unsigned and signed.
	switch pred {
	case ir.IntEQ:
		return token.EQL
	case ir.IntNE:
		return token.NEQ
	case ir.IntUGT:
		return token.GTR
	case ir.IntUGE:
		return token.GEQ
	case ir.IntULT:
		return token.LSS
	case ir.IntULE:
		return token.LEQ
	case ir.IntSGT:
		return token.GTR
	case ir.IntSGE:
		return token.GEQ
	case ir.IntSLT:
		return token.LSS
	case ir.IntSLE:
		return token.LEQ
	default:
		panic(fmt.Sprintf("support for integer predicate %v not yet implemented", pred))
	}
}

// floatPred converts the given LLVM IR floating-point predicate to a
// corresponding Go token.
func floatPred(pred ir.FloatPred) token.Token {
	// TODO: Differentiate between ordered and unordered.
	switch pred {
	case ir.FloatFalse:
		panic(`support for floating-point predicate "false" not yet implemented`)
	case ir.FloatOEQ:
		return token.EQL
	case ir.FloatOGT:
		return token.GTR
	case ir.FloatOGE:
		return token.GEQ
	case ir.FloatOLT:
		return token.LSS
	case ir.FloatOLE:
		return token.LEQ
	case ir.FloatONE:
		return token.NEQ
	case ir.FloatORD:
		panic(`support for floating-point predicate "ord" not yet implemented`)
	case ir.FloatUEQ:
		return token.EQL
	case ir.FloatUGT:
		return token.GTR
	case ir.FloatUGE:
		return token.GEQ
	case ir.FloatULT:
		return token.LSS
	case ir.FloatULE:
		return token.LEQ
	case ir.FloatUNE:
		return token.NEQ
	case ir.FloatUNO:
		panic(`support for floating-point predicate "uno" not yet implemented`)
	case ir.FloatTrue:
		panic(`support for floating-point predicate "true" not yet implemented`)
	default:
		panic(fmt.Sprintf("support for floating-point predicate %v not yet implemented", pred))
	}
}
