// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/token"
	"log"
)

func init() {
	register(assignbinopFix)
}

var assignbinopFix = fix{
	"assignbinop",
	"2015-03-11",
	assignbinop,
	`Replace "x = x + z" with "x += z".`,
}

func assignbinop(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    x = x + y
	//
	//    // to:
	//    x += y
	//
	// 2)
	//    // from:
	//    x = y * x
	//
	//    // to:
	//    x *= y
	walk(file, func(n interface{}) {
		stmt, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		assignStmt, ok := (*stmt).(*ast.AssignStmt)
		if !ok {
			return
		}
		if assignStmt.Tok != token.ASSIGN {
			return
		}
		lhs := assignStmt.Lhs
		if len(lhs) != 1 {
			return
		}
		ident, ok := lhs[0].(*ast.Ident)
		if !ok {
			return
		}
		rhs := assignStmt.Rhs
		binExpr, ok := rhs[0].(*ast.BinaryExpr)
		if !ok {
			return
		}
		x, y := binExpr.X, binExpr.Y
		one := false
		switch {
		case isName(x, ident.Name):
			// x = x + y
			one = isOne(y)
			rhs = []ast.Expr{y}
		case isName(y, ident.Name):
			// x = y + x
			one = isOne(x)
			rhs = []ast.Expr{x}
		default:
			return
		}
		var op token.Token
		switch binExpr.Op {
		case token.ADD:
			op = token.ADD_ASSIGN // +=
			if one {
				// x++
				*stmt = &ast.IncDecStmt{
					X:   ident,
					Tok: token.INC,
				}
				fixed = true
				return
			}
		case token.SUB:
			op = token.SUB_ASSIGN // -=
			if one {
				// x--
				*stmt = &ast.IncDecStmt{
					X:   ident,
					Tok: token.DEC,
				}
				fixed = true
				return
			}
		case token.MUL:
			op = token.MUL_ASSIGN // *=
		case token.QUO:
			op = token.QUO_ASSIGN // /=
		case token.REM:
			op = token.REM_ASSIGN // %=
		case token.AND:
			op = token.AND_ASSIGN // &=
		case token.OR:
			op = token.OR_ASSIGN // |=
		case token.XOR:
			op = token.XOR_ASSIGN // ^=
		case token.SHL:
			op = token.SHL_ASSIGN // <<=
		case token.SHR:
			op = token.SHR_ASSIGN // >>=
		case token.AND_NOT:
			op = token.AND_NOT_ASSIGN // &^=
		default:
			log.Fatalf("unknown binary operand %v\n", binExpr.Op)
		}
		*stmt = &ast.AssignStmt{
			Lhs: lhs,
			Tok: op,
			Rhs: rhs,
		}
		fixed = true
	})

	return fixed
}

// isOne returns true if n is the integer literal 1, and false otherwise.
func isOne(n ast.Expr) bool {
	lit, ok := n.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == "1"
}
