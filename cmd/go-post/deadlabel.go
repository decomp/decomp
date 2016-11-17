// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "go/ast"

func init() {
	register(deadlabelFix)
}

var deadlabelFix = fix{
	"deadlabel",
	"2016-11-17",
	deadlabel,
	`Remove unused labels.`,
}

func deadlabel(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    foo:
	//       sum := 0
	//    loop:
	//       for i := 0; i < 10; i++ {
	//    bar:
	//          sum += i
	//    baz:
	//          if sum > 10 {
	//    qux:
	//             break loop
	//          }
	//       }
	//
	//    // to:
	//       sum := 0
	//    loop:
	//       for i := 0; i < 10; i++ {
	//          sum += i
	//          if sum > 10 {
	//             break loop
	//          }
	//       }

	for _, decl := range file.Decls {
		f, ok := decl.(*ast.FuncDecl)
		if !ok || f.Body == nil {
			continue
		}
		// Identify used labels.
		used := make(map[string]bool)
		walk(f.Body, func(n interface{}) {
			branch, ok := n.(*ast.BranchStmt)
			if !ok {
				return
			}
			used[branch.Label.String()] = true
		})
		// Remove unused labels.
		walk(f.Body, func(n interface{}) {
			stmt, ok := n.(*ast.Stmt)
			if !ok {
				return
			}
			labeledStmt, ok := (*stmt).(*ast.LabeledStmt)
			if !ok {
				return
			}
			if !used[labeledStmt.Label.String()] {
				// Remove label.
				*stmt = labeledStmt.Stmt
				fixed = true
			}
		})
	}

	return fixed
}
