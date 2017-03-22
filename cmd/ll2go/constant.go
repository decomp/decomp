package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/llir/llvm/ir/constant"
)

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
	panic("not yet implemented")
}

// exprFAdd converts the given LLVM IR fadd expression to a corresponding Go
// statement.
func (d *decompiler) exprFAdd(expr *constant.ExprFAdd) ast.Expr {
	panic("not yet implemented")
}

// exprSub converts the given LLVM IR sub expression to a corresponding Go
// statement.
func (d *decompiler) exprSub(expr *constant.ExprSub) ast.Expr {
	panic("not yet implemented")
}

// exprFSub converts the given LLVM IR fsub expression to a corresponding Go
// statement.
func (d *decompiler) exprFSub(expr *constant.ExprFSub) ast.Expr {
	panic("not yet implemented")
}

// exprMul converts the given LLVM IR mul expression to a corresponding Go
// statement.
func (d *decompiler) exprMul(expr *constant.ExprMul) ast.Expr {
	panic("not yet implemented")
}

// exprFMul converts the given LLVM IR fmul expression to a corresponding Go
// statement.
func (d *decompiler) exprFMul(expr *constant.ExprFMul) ast.Expr {
	panic("not yet implemented")
}

// exprUDiv converts the given LLVM IR udiv expression to a corresponding Go
// statement.
func (d *decompiler) exprUDiv(expr *constant.ExprUDiv) ast.Expr {
	panic("not yet implemented")
}

// exprSDiv converts the given LLVM IR sdiv expression to a corresponding Go
// statement.
func (d *decompiler) exprSDiv(expr *constant.ExprSDiv) ast.Expr {
	panic("not yet implemented")
}

// exprFDiv converts the given LLVM IR fdiv expression to a corresponding Go
// statement.
func (d *decompiler) exprFDiv(expr *constant.ExprFDiv) ast.Expr {
	panic("not yet implemented")
}

// exprURem converts the given LLVM IR urem expression to a corresponding Go
// statement.
func (d *decompiler) exprURem(expr *constant.ExprURem) ast.Expr {
	panic("not yet implemented")
}

// exprSRem converts the given LLVM IR srem expression to a corresponding Go
// statement.
func (d *decompiler) exprSRem(expr *constant.ExprSRem) ast.Expr {
	panic("not yet implemented")
}

// exprFRem converts the given LLVM IR frem expression to a corresponding Go
// statement.
func (d *decompiler) exprFRem(expr *constant.ExprFRem) ast.Expr {
	panic("not yet implemented")
}

// exprShl converts the given LLVM IR shl expression to a corresponding Go
// statement.
func (d *decompiler) exprShl(expr *constant.ExprShl) ast.Expr {
	panic("not yet implemented")
}

// exprLShr converts the given LLVM IR lshr expression to a corresponding Go
// statement.
func (d *decompiler) exprLShr(expr *constant.ExprLShr) ast.Expr {
	panic("not yet implemented")
}

// exprAShr converts the given LLVM IR ashr expression to a corresponding Go
// statement.
func (d *decompiler) exprAShr(expr *constant.ExprAShr) ast.Expr {
	panic("not yet implemented")
}

// exprAnd converts the given LLVM IR and expression to a corresponding Go
// statement.
func (d *decompiler) exprAnd(expr *constant.ExprAnd) ast.Expr {
	panic("not yet implemented")
}

// exprOr converts the given LLVM IR or expression to a corresponding Go
// statement.
func (d *decompiler) exprOr(expr *constant.ExprOr) ast.Expr {
	panic("not yet implemented")
}

// exprXor converts the given LLVM IR xor expression to a corresponding Go
// statement.
func (d *decompiler) exprXor(expr *constant.ExprXor) ast.Expr {
	panic("not yet implemented")
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
	panic("not yet implemented")
}

// exprZExt converts the given LLVM IR zext expression to a corresponding Go
// statement.
func (d *decompiler) exprZExt(expr *constant.ExprZExt) ast.Expr {
	panic("not yet implemented")
}

// exprSExt converts the given LLVM IR sext expression to a corresponding Go
// statement.
func (d *decompiler) exprSExt(expr *constant.ExprSExt) ast.Expr {
	panic("not yet implemented")
}

// exprFPTrunc converts the given LLVM IR fptrunc expression to a corresponding
// Go statement.
func (d *decompiler) exprFPTrunc(expr *constant.ExprFPTrunc) ast.Expr {
	panic("not yet implemented")
}

// exprFPExt converts the given LLVM IR fpext expression to a corresponding Go
// statement.
func (d *decompiler) exprFPExt(expr *constant.ExprFPExt) ast.Expr {
	panic("not yet implemented")
}

// exprFPToUI converts the given LLVM IR fptoui expression to a corresponding Go
// statement.
func (d *decompiler) exprFPToUI(expr *constant.ExprFPToUI) ast.Expr {
	panic("not yet implemented")
}

// exprFPToSI converts the given LLVM IR fptosi expression to a corresponding Go
// statement.
func (d *decompiler) exprFPToSI(expr *constant.ExprFPToSI) ast.Expr {
	panic("not yet implemented")
}

// exprUIToFP converts the given LLVM IR uitofp expression to a corresponding Go
// statement.
func (d *decompiler) exprUIToFP(expr *constant.ExprUIToFP) ast.Expr {
	panic("not yet implemented")
}

// exprSIToFP converts the given LLVM IR sitofp expression to a corresponding Go
// statement.
func (d *decompiler) exprSIToFP(expr *constant.ExprSIToFP) ast.Expr {
	panic("not yet implemented")
}

// exprPtrToInt converts the given LLVM IR ptrtoint expression to a
// corresponding Go statement.
func (d *decompiler) exprPtrToInt(expr *constant.ExprPtrToInt) ast.Expr {
	panic("not yet implemented")
}

// exprIntToPtr converts the given LLVM IR inttoptr expression to a
// corresponding Go statement.
func (d *decompiler) exprIntToPtr(expr *constant.ExprIntToPtr) ast.Expr {
	panic("not yet implemented")
}

// exprBitCast converts the given LLVM IR bitcast expression to a corresponding
// Go statement.
func (d *decompiler) exprBitCast(expr *constant.ExprBitCast) ast.Expr {
	panic("not yet implemented")
}

// exprAddrSpaceCast converts the given LLVM IR addrspacecast expression to a
// corresponding Go statement.
func (d *decompiler) exprAddrSpaceCast(expr *constant.ExprAddrSpaceCast) ast.Expr {
	panic("not yet implemented")
}

// exprICmp converts the given LLVM IR icmp expression to a corresponding Go
// statement.
func (d *decompiler) exprICmp(expr *constant.ExprICmp) ast.Expr {
	panic("not yet implemented")
}

// exprFCmp converts the given LLVM IR fcmp expression to a corresponding Go
// statement.
func (d *decompiler) exprFCmp(expr *constant.ExprFCmp) ast.Expr {
	panic("not yet implemented")
}

// exprSelect converts the given LLVM IR select expression to a corresponding Go
// statement.
func (d *decompiler) exprSelect(expr *constant.ExprSelect) ast.Expr {
	panic("not yet implemented")
}
