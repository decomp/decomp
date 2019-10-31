package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(deadassignFix)
}

var deadassignFix = fix{
	name:     "deadassign",
	date:     "2015-03-11",
	f:        deadassign,
	desc:     `Remove "x = x" assignments.`,
	disabled: false,
}

func deadassign(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    a := 1
	//    x = x
	//
	//    // to:
	//    a := 1
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
		if len(assignStmt.Lhs) != 1 {
			return
		}
		lid, ok := assignStmt.Lhs[0].(*ast.Ident)
		if !ok {
			return
		}
		rid, ok := assignStmt.Rhs[0].(*ast.Ident)
		if !ok {
			return
		}
		if isName(lid, rid.Name) {
			*stmt = &ast.EmptyStmt{}
			fixed = true
		}
	})

	return fixed
}
