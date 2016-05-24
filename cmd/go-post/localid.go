// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func init() {
	register(localidFix)
}

var localidFix = fix{
	"localid",
	// HACK: Fixes are sorted by date. The Unix epoch makes sure that the local
	// ID replacement rule happens before all other rules. This enables
	// assignbinop simplification directly.
	"1970-01-01",
	localid,
	`Replace the use of local variable IDs with their definition.`,
}

func localid(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    _0 = i < 10
	//    if _0 {}
	//
	//    // to:
	//    _0 = i < 10
	//    if i < 10 {}
	//
	// 2)
	//    // from:
	//    _0 = i + j
	//    _1 = x * y
	//    a := _0 + _1
	//
	//    // to:
	//    a := (i + j) + (x * y)
	walk(file, func(n interface{}) {
		stmt, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		assignStmt, ok := (*stmt).(*ast.AssignStmt)
		if !ok {
			return
		}
		if len(assignStmt.Lhs) != 1 {
			return
		}
		ident, ok := assignStmt.Lhs[0].(*ast.Ident)
		if !ok {
			return
		}
		if name := ident.Name; isLocalID(name) {
			rhs := assignStmt.Rhs[0]
			// TODO: Make use of &ast.ParenExpr{} and implement a simplification
			// pass which takes operator precedence into account.
			f := func(pos token.Pos) ast.Expr {
				fixed = true
				return rhs
				//return &ast.ParenExpr{X: rhs}
			}
			fnot := func(pos token.Pos) ast.Expr {
				fixed = true
				return &ast.UnaryExpr{
					OpPos: pos,
					Op:    token.NOT,
					X:     rhs,
					//X:     &ast.ParenExpr{X: rhs},
				}
			}
			scope := getScope(file, ident)
			rewriteUses(ident, f, fnot, scope)
			*stmt = &ast.EmptyStmt{}
		}
	})

	return fixed
}

// getScope returns all statements in which ident is in scope.
func getScope(file *ast.File, ident *ast.Ident) []ast.Stmt {
	var scope []ast.Stmt
	f := func(n interface{}) {
		stmt, ok := n.(ast.Stmt)
		if !ok {
			return
		}
		if countUses(ident, []ast.Stmt{stmt}) > 0 {
			scope = append(scope, stmt)
		}
	}
	walk(file, f)
	return scope
}

// isLocalID returns true if the given variable name is a local ID (e.g. "_42").
func isLocalID(name string) bool {
	if strings.HasPrefix(name, "_") {
		_, err := strconv.Atoi(name[1:])
		if err == nil {
			return true
		}
	}
	return false
}
