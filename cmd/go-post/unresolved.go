// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(unresolvedFix)
}

var unresolvedFix = fix{
	"unresolved",
	"2015-03-11",
	unresolved,
	`Replace assignment statements with declare and initialize statements at the first occurance of an unresolved identifier.`,
}

func unresolved(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    i = 20 // undefined: i
	//
	//    // to:
	//    i := 20
	walk(file, func(n interface{}) {
		// Early return if already fixed. The next iteration of go fix will have
		// updated file.Unresolved for us.
		if fixed {
			return
		}

		stmt, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		assignStmt, ok := (*stmt).(*ast.AssignStmt)
		if !ok {
			return
		}
		for _, expr := range assignStmt.Lhs {
			ident, ok := expr.(*ast.Ident)
			if !ok {
				continue
			}
			if isUnresolved(file, ident) {
				// Replace "=" token with ":="
				assignStmt.Tok = token.DEFINE
				fixed = true
				break
			}
		}
	})

	return fixed
}

// isUnresolved returns true if the given identifier is unresolved, and false
// otherwise.
func isUnresolved(file *ast.File, ident *ast.Ident) bool {
	for _, u := range file.Unresolved {
		if u == ident {
			return true
		}
	}
	return false
}
