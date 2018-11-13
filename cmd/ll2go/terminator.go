package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/llir/llvm/ir"
)

// term converts the given LLVM IR terminator to a corresponding Go statement.
func (d *decompiler) term(term ir.Terminator) ast.Stmt {
	switch term := term.(type) {
	case *ir.TermRet:
		return d.termRet(term)
	case *ir.TermBr:
		return d.termBr(term)
	case *ir.TermCondBr:
		return d.termCondBr(term)
	case *ir.TermSwitch:
		return d.termSwitch(term)
	case *ir.TermUnreachable:
		return d.termUnreachable(term)
	default:
		panic(fmt.Sprintf("support for terminator %T not yet implemented", term))
	}
}

// termRet converts the given LLVM IR ret terminator to a corresponding Go
// statement.
func (d *decompiler) termRet(term *ir.TermRet) ast.Stmt {
	if term.X == nil {
		return &ast.ReturnStmt{}
	}
	return &ast.ReturnStmt{
		Results: []ast.Expr{d.value(term.X)},
	}
}

// termBr converts the given LLVM IR br terminator to a corresponding Go
// statement.
func (d *decompiler) termBr(term *ir.TermBr) ast.Stmt {
	// Use goto-statements as a fallback for incomplete control flow recovery.
	d.labels[term.Target.LocalName] = true
	return &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: d.label(term.Target.LocalName),
	}
}

// termCondBr converts the given LLVM IR conditional br terminator to a
// corresponding Go statement.
func (d *decompiler) termCondBr(term *ir.TermCondBr) ast.Stmt {
	// Use goto-statements as a fallback for incomplete control flow recovery.
	d.labels[term.TargetTrue.LocalName] = true
	d.labels[term.TargetFalse.LocalName] = true
	gotoTrueStmt := &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: d.label(term.TargetTrue.LocalName),
	}
	gotoFalseStmt := &ast.BranchStmt{
		Tok:   token.GOTO,
		Label: d.label(term.TargetFalse.LocalName),
	}
	return &ast.IfStmt{
		Cond: d.value(term.Cond),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{gotoTrueStmt},
		},
		Else: &ast.BlockStmt{
			List: []ast.Stmt{gotoFalseStmt},
		},
	}
}

// termSwitch converts the given LLVM IR switch terminator to a corresponding Go
// statement.
func (d *decompiler) termSwitch(term *ir.TermSwitch) ast.Stmt {
	// Use goto-statements as a fallback for incomplete control flow recovery.
	var cases []ast.Stmt
	for _, c := range term.Cases {
		gotoStmt := &ast.BranchStmt{
			Tok:   token.GOTO,
			Label: d.localIdent(c.Target.LocalName),
		}
		cc := &ast.CaseClause{
			List: []ast.Expr{d.value(c.X)},
			Body: []ast.Stmt{gotoStmt},
		}
		cases = append(cases, cc)
	}
	return &ast.SwitchStmt{
		Tag: d.value(term.X),
		Body: &ast.BlockStmt{
			List: cases,
		},
	}
}

// termUnreachable converts the given LLVM IR unreachable terminator to a
// corresponding Go statement.
func (d *decompiler) termUnreachable(term *ir.TermUnreachable) ast.Stmt {
	unreachable := &ast.BasicLit{
		Kind:  token.STRING,
		Value: `"unreachable"`,
	}
	expr := &ast.CallExpr{
		Fun:  ast.NewIdent("panic"),
		Args: []ast.Expr{unreachable},
	}
	return &ast.ExprStmt{
		X: expr,
	}
}
