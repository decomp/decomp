package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
)

// constant converts the given LLVM IR constant to a corresponding Go
// expression.
func (d *decompiler) constant(c constant.Constant) ast.Expr {
	switch c := c.(type) {
	// Simple constants
	case *constant.Int:
		return d.constInt(c)
	case *constant.Float:
		return d.constFloat(c)
	case *constant.Null:
		return d.constNull(c)
	// Complex constants
	case *constant.Vector:
		panic("support for *constant.Vector not yet implemented")
	case *constant.Array:
		return d.constArray(c)
	case *constant.Struct:
		panic("support for *constant.Struct not yet implemented")
	case *constant.ZeroInitializer:
		panic("support for *constant.ZeroInitializer not yet implemented")
	// Global variable and function addresses
	case *ir.Global:
		return d.global(c.Name)
	case *ir.Function:
		return d.global(c.Name)
	// Constant expressions
	case constant.Expr:
		return d.expr(c)
	default:
		panic(fmt.Sprintf("support for constant value %T not yet implemented", c))
	}
}

// constInt converts the given LLVM IR integer constant to a corresponding Go
// expression.
func (d *decompiler) constInt(c *constant.Int) ast.Expr {
	return &ast.BasicLit{
		Kind:  token.INT,
		Value: c.X.String(),
	}
}

// constFloat converts the given LLVM IR floating-point constant to a
// corresponding Go expression.
func (d *decompiler) constFloat(c *constant.Float) ast.Expr {
	return &ast.BasicLit{
		Kind:  token.FLOAT,
		Value: c.X.String(),
	}
}

// constNull converts the given LLVM IR null pointer constant constant to a
// corresponding Go expression.
func (d *decompiler) constNull(c *constant.Null) ast.Expr {
	return ast.NewIdent("nil")
}

// constArray converts the given LLVM IR array constant constant to a
// corresponding Go expression.
func (d *decompiler) constArray(c *constant.Array) ast.Expr {
	var elems []ast.Expr
	for _, e := range c.Elems {
		elems = append(elems, d.value(e))
	}
	return &ast.CompositeLit{
		Type: d.goType(c.Typ),
		Elts: elems,
	}
}

// expr converts the given LLVM IR expression to a corresponding Go expression.
func (d *decompiler) expr(expr constant.Expr) ast.Expr {
	switch expr := expr.(type) {
	// Binary expressions
	case *constant.ExprAdd:
		return d.exprAdd(expr)
	case *constant.ExprFAdd:
		return d.exprFAdd(expr)
	case *constant.ExprSub:
		return d.exprSub(expr)
	case *constant.ExprFSub:
		return d.exprFSub(expr)
	case *constant.ExprMul:
		return d.exprMul(expr)
	case *constant.ExprFMul:
		return d.exprFMul(expr)
	case *constant.ExprUDiv:
		return d.exprUDiv(expr)
	case *constant.ExprSDiv:
		return d.exprSDiv(expr)
	case *constant.ExprFDiv:
		return d.exprFDiv(expr)
	case *constant.ExprURem:
		return d.exprURem(expr)
	case *constant.ExprSRem:
		return d.exprSRem(expr)
	case *constant.ExprFRem:
		return d.exprFRem(expr)
	// Bitwise expressions
	case *constant.ExprShl:
		return d.exprShl(expr)
	case *constant.ExprLShr:
		return d.exprLShr(expr)
	case *constant.ExprAShr:
		return d.exprAShr(expr)
	case *constant.ExprAnd:
		return d.exprAnd(expr)
	case *constant.ExprOr:
		return d.exprOr(expr)
	case *constant.ExprXor:
		return d.exprXor(expr)
	// Memory expressions
	case *constant.ExprGetElementPtr:
		return d.exprGetElementPtr(expr)
	// Conversion expressions
	case *constant.ExprTrunc:
		return d.exprTrunc(expr)
	case *constant.ExprZExt:
		return d.exprZExt(expr)
	case *constant.ExprSExt:
		return d.exprSExt(expr)
	case *constant.ExprFPTrunc:
		return d.exprFPTrunc(expr)
	case *constant.ExprFPExt:
		return d.exprFPExt(expr)
	case *constant.ExprFPToUI:
		return d.exprFPToUI(expr)
	case *constant.ExprFPToSI:
		return d.exprFPToSI(expr)
	case *constant.ExprUIToFP:
		return d.exprUIToFP(expr)
	case *constant.ExprSIToFP:
		return d.exprSIToFP(expr)
	case *constant.ExprPtrToInt:
		return d.exprPtrToInt(expr)
	case *constant.ExprIntToPtr:
		return d.exprIntToPtr(expr)
	case *constant.ExprBitCast:
		return d.exprBitCast(expr)
	case *constant.ExprAddrSpaceCast:
		return d.exprAddrSpaceCast(expr)
	// Other expressions
	case *constant.ExprICmp:
		return d.exprICmp(expr)
	case *constant.ExprFCmp:
		return d.exprFCmp(expr)
	case *constant.ExprSelect:
		return d.exprSelect(expr)
	default:
		panic(fmt.Sprintf("support for expression %T not yet implemented", expr))
	}
}

// exprAdd converts the given LLVM IR add expression to a corresponding Go
// statement.
func (d *decompiler) exprAdd(expr *constant.ExprAdd) ast.Expr {
	return d.binaryOp(expr.X, token.ADD, expr.Y)
}

// exprFAdd converts the given LLVM IR fadd expression to a corresponding Go
// statement.
func (d *decompiler) exprFAdd(expr *constant.ExprFAdd) ast.Expr {
	return d.binaryOp(expr.X, token.ADD, expr.Y)
}

// exprSub converts the given LLVM IR sub expression to a corresponding Go
// statement.
func (d *decompiler) exprSub(expr *constant.ExprSub) ast.Expr {
	return d.binaryOp(expr.X, token.SUB, expr.Y)
}

// exprFSub converts the given LLVM IR fsub expression to a corresponding Go
// statement.
func (d *decompiler) exprFSub(expr *constant.ExprFSub) ast.Expr {
	return d.binaryOp(expr.X, token.SUB, expr.Y)
}

// exprMul converts the given LLVM IR mul expression to a corresponding Go
// statement.
func (d *decompiler) exprMul(expr *constant.ExprMul) ast.Expr {
	return d.binaryOp(expr.X, token.MUL, expr.Y)
}

// exprFMul converts the given LLVM IR fmul expression to a corresponding Go
// statement.
func (d *decompiler) exprFMul(expr *constant.ExprFMul) ast.Expr {
	return d.binaryOp(expr.X, token.MUL, expr.Y)
}

// exprUDiv converts the given LLVM IR udiv expression to a corresponding Go
// statement.
func (d *decompiler) exprUDiv(expr *constant.ExprUDiv) ast.Expr {
	return d.binaryOp(expr.X, token.QUO, expr.Y)
}

// exprSDiv converts the given LLVM IR sdiv expression to a corresponding Go
// statement.
func (d *decompiler) exprSDiv(expr *constant.ExprSDiv) ast.Expr {
	return d.binaryOp(expr.X, token.QUO, expr.Y)
}

// exprFDiv converts the given LLVM IR fdiv expression to a corresponding Go
// statement.
func (d *decompiler) exprFDiv(expr *constant.ExprFDiv) ast.Expr {
	return d.binaryOp(expr.X, token.QUO, expr.Y)
}

// exprURem converts the given LLVM IR urem expression to a corresponding Go
// statement.
func (d *decompiler) exprURem(expr *constant.ExprURem) ast.Expr {
	return d.binaryOp(expr.X, token.REM, expr.Y)
}

// exprSRem converts the given LLVM IR srem expression to a corresponding Go
// statement.
func (d *decompiler) exprSRem(expr *constant.ExprSRem) ast.Expr {
	return d.binaryOp(expr.X, token.REM, expr.Y)
}

// exprFRem converts the given LLVM IR frem expression to a corresponding Go
// statement.
func (d *decompiler) exprFRem(expr *constant.ExprFRem) ast.Expr {
	return d.binaryOp(expr.X, token.REM, expr.Y)
}

// exprShl converts the given LLVM IR shl expression to a corresponding Go
// statement.
func (d *decompiler) exprShl(expr *constant.ExprShl) ast.Expr {
	return d.binaryOp(expr.X, token.SHL, expr.Y)
}

// exprLShr converts the given LLVM IR lshr expression to a corresponding Go
// statement.
func (d *decompiler) exprLShr(expr *constant.ExprLShr) ast.Expr {
	return d.binaryOp(expr.X, token.SHR, expr.Y)
}

// exprAShr converts the given LLVM IR ashr expression to a corresponding Go
// statement.
func (d *decompiler) exprAShr(expr *constant.ExprAShr) ast.Expr {
	return d.binaryOp(expr.X, token.SHR, expr.Y)
}

// exprAnd converts the given LLVM IR and expression to a corresponding Go
// statement.
func (d *decompiler) exprAnd(expr *constant.ExprAnd) ast.Expr {
	return d.binaryOp(expr.X, token.AND, expr.Y)
}

// exprOr converts the given LLVM IR or expression to a corresponding Go
// statement.
func (d *decompiler) exprOr(expr *constant.ExprOr) ast.Expr {
	return d.binaryOp(expr.X, token.OR, expr.Y)
}

// exprXor converts the given LLVM IR xor expression to a corresponding Go
// statement.
func (d *decompiler) exprXor(expr *constant.ExprXor) ast.Expr {
	return d.binaryOp(expr.X, token.XOR, expr.Y)
}

// exprGetElementPtr converts the given LLVM IR getelementptr expression to a
// corresponding Go statement.
func (d *decompiler) exprGetElementPtr(expr *constant.ExprGetElementPtr) ast.Expr {
	src := d.value(expr.Src)
	// TODO: Validate if index expressions should be added in reverse order.
	for _, index := range expr.Indices {
		src = &ast.IndexExpr{
			X:     src,
			Index: d.value(index),
		}
	}
	e := &ast.UnaryExpr{
		Op: token.AND,
		X:  src,
	}
	return e
}

// exprTrunc converts the given LLVM IR trunc expression to a corresponding Go
// statement.
func (d *decompiler) exprTrunc(expr *constant.ExprTrunc) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprZExt converts the given LLVM IR zext expression to a corresponding Go
// statement.
func (d *decompiler) exprZExt(expr *constant.ExprZExt) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprSExt converts the given LLVM IR sext expression to a corresponding Go
// statement.
func (d *decompiler) exprSExt(expr *constant.ExprSExt) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprFPTrunc converts the given LLVM IR fptrunc expression to a corresponding
// Go statement.
func (d *decompiler) exprFPTrunc(expr *constant.ExprFPTrunc) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprFPExt converts the given LLVM IR fpext expression to a corresponding Go
// statement.
func (d *decompiler) exprFPExt(expr *constant.ExprFPExt) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprFPToUI converts the given LLVM IR fptoui expression to a corresponding Go
// statement.
func (d *decompiler) exprFPToUI(expr *constant.ExprFPToUI) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprFPToSI converts the given LLVM IR fptosi expression to a corresponding Go
// statement.
func (d *decompiler) exprFPToSI(expr *constant.ExprFPToSI) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprUIToFP converts the given LLVM IR uitofp expression to a corresponding Go
// statement.
func (d *decompiler) exprUIToFP(expr *constant.ExprUIToFP) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprSIToFP converts the given LLVM IR sitofp expression to a corresponding Go
// statement.
func (d *decompiler) exprSIToFP(expr *constant.ExprSIToFP) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprPtrToInt converts the given LLVM IR ptrtoint expression to a
// corresponding Go statement.
func (d *decompiler) exprPtrToInt(expr *constant.ExprPtrToInt) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprIntToPtr converts the given LLVM IR inttoptr expression to a
// corresponding Go statement.
func (d *decompiler) exprIntToPtr(expr *constant.ExprIntToPtr) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprBitCast converts the given LLVM IR bitcast expression to a corresponding
// Go statement.
func (d *decompiler) exprBitCast(expr *constant.ExprBitCast) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprAddrSpaceCast converts the given LLVM IR addrspacecast expression to a
// corresponding Go statement.
func (d *decompiler) exprAddrSpaceCast(expr *constant.ExprAddrSpaceCast) ast.Expr {
	// TODO: Add proper support for type conversion. For now, simply return the
	// original value.
	return d.value(expr.From)
}

// exprICmp converts the given LLVM IR icmp expression to a corresponding Go
// statement.
func (d *decompiler) exprICmp(expr *constant.ExprICmp) ast.Expr {
	op := intPred(ir.IntPred(expr.Pred))
	return d.binaryOp(expr.X, op, expr.Y)
}

// exprFCmp converts the given LLVM IR fcmp expression to a corresponding Go
// statement.
func (d *decompiler) exprFCmp(expr *constant.ExprFCmp) ast.Expr {
	op := floatPred(ir.FloatPred(expr.Pred))
	return d.binaryOp(expr.X, op, expr.Y)
}

// exprSelect converts the given LLVM IR select expression to a corresponding Go
// statement.
func (d *decompiler) exprSelect(expr *constant.ExprSelect) ast.Expr {
	panic("not yet implemented")
}
