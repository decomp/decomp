// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(forloopFix)
}

var forloopFix = fix{
	"forloop",
	// NOTE: The "forloop" go fix rule depends on the "unresolved" go fix rule as
	// it locates undeclared variables and declares them using ":=". The date of
	// "forloop" must therefore be later in time than the date of "unresolved".
	"2015-03-12",
	forloop,
	`Add initialization and post-statements to for-loops.`,
}

func forloop(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    i := 0
	//    for i < 10 {
	//       i++
	//    }
	//
	//    // to:
	//    for i := 0; i < 10; i++ {
	//    }
	walk(file, func(n interface{}) {
		blockStmt, ok := n.(**ast.BlockStmt)
		if !ok {
			return
		}
		list := (*blockStmt).List
		for i, stmt0 := range list {
			forStmt, ok := stmt0.(*ast.ForStmt)
			if !ok {
				continue
			}
			// for init; cond; post {}
			initStmt, condStmt, postStmt := forStmt.Init, forStmt.Cond, forStmt.Post
			if initStmt != nil || postStmt != nil {
				continue
			}
			// x < 10
			cond, ok := condStmt.(*ast.BinaryExpr)
			if !ok {
				continue
			}
			switch cond.Op {
			case token.EQL, token.GEQ, token.GTR, token.LEQ, token.LSS, token.NEQ:
				// ==, >=, >, <=, < or !=
			default:
				continue
			}
			x, ok := cond.X.(*ast.Ident)
			if !ok {
				continue
			}
			initPos := initIndex(list[:i], x)
			if initPos == -1 {
				continue
			}
			postPos := postIndex(forStmt.Body.List, x)
			if initPos == -1 {
				continue
			}
			// Remove "i := 0" from the block statement list.
			forStmt.Init = list[initPos]
			(*blockStmt).List = listDel(list, initPos)
			// Remove "i++" from the for-body.
			forStmt.Post = forStmt.Body.List[postPos]
			forStmt.Body.List = listDel(forStmt.Body.List, postPos)
			fixed = true
		}
	})

	return fixed
}

// listDel removes the i:th statement of list.
func listDel(list []ast.Stmt, i int) []ast.Stmt {
	return append(list[:i], list[i+1:]...)
}

// initIndex returns the position of the statement which defines x in list, or
// -1 if no such declaration could be located.
func initIndex(list []ast.Stmt, x *ast.Ident) (pos int) {
	for i := len(list) - 1; i >= 0; i-- {
		stmt := list[i]
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		// TODO: Include support for "var" declarations.
		if assign.Tok != token.DEFINE {
			continue
		}
		lhs := assign.Lhs
		if len(lhs) != 1 {
			continue
		}
		if isName(lhs[0], x.Name) {
			return i
		}
	}
	return -1
}

// TODO: Add support for arbitrary binary assign ops (e.g. +=, -=, /=, etc).

// postIndex returns the position of the last increment or decrement of x in
// list, or -1 if no such statement could be located.
func postIndex(list []ast.Stmt, x *ast.Ident) (pos int) {
	// TODO: Add sanity check which makes sure "x" is not used by any statement
	// after x++.
	for i := len(list) - 1; i >= 0; i-- {
		stmt := list[i]
		incdec, ok := stmt.(*ast.IncDecStmt)
		if !ok {
			continue
		}
		if isName(incdec.X, x.Name) {
			return i
		}
	}
	return -1
}
