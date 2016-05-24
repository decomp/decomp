// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(deadassignFix)
}

var deadassignFix = fix{
	"deadassign",
	"2015-03-11",
	deadassign,
	`Remove "x = x" assignments.`,
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
