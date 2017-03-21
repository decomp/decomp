package main

import (
	"fmt"
	"go/ast"

	"github.com/llir/llvm/ir"
)

// term converts the given LLVM IR terminator to a corresponding Go statement.
func (d *decompiler) term(term ir.Terminator) ast.Stmt {
	switch term := term.(type) {
	case *ir.TermRet:
		var results []ast.Expr
		if term.X != nil {
			results = append(results, d.value(term.X))
		}
		return &ast.ReturnStmt{
			Results: results,
		}
	case *ir.TermBr:
		panic("support for terminator *ir.TermBr not yet implemented")
	case *ir.TermCondBr:
		panic("support for terminator *ir.TermCondBr not yet implemented")
	case *ir.TermSwitch:
		panic("support for terminator *ir.TermSwitch not yet implemented")
	case *ir.TermUnreachable:
		panic("support for terminator *ir.TermUnreachable not yet implemented")
	default:
		panic(fmt.Sprintf("support for terminator %T not yet implemented", term))
	}
}
