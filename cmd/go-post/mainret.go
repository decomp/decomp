// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

func init() {
	register(mainretFix)
}

var mainretFix = fix{
	"mainret",
	"2015-02-27",
	mainret,
	`Replace return statements with calls to os.Exit in the "main" function.`,
}

func mainret(file *ast.File) bool {
	fixed := false

	// Add "os" import.
	addImport(file, "os")

	// Locate the "main" function.
	mainFunc, ok := findMainFunc(file)
	if !ok {
		return false
	}

	// Apply the following transitions for the "main" function:
	//
	// 1)
	//    // from:
	//    return 42
	//
	//    // to:
	//    os.Exit(42)
	//
	// 2)
	//    // from:
	//    return 0
	//
	//    // to:
	//    return
	//
	// 3)
	//    // from:
	//    func main() {
	//       return
	//    }
	//
	//    // to:
	//    func main() {
	//    }
	walk(mainFunc, func(n interface{}) {
		stmt, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		retStmt, ok := (*stmt).(*ast.ReturnStmt)
		if !ok {
			return
		}
		switch len(retStmt.Results) {
		case 0:
			// Leave blank returns as is.
			return
		case 1:
			result := retStmt.Results[0]
			if isZero(result) {
				// Replace "return 0" with "return".
				retStmt.Results = nil
			} else {
				// Replace "return 42" with "os.Exit(42)".
				exit := createExit(result)
				*stmt = exit
			}
			fixed = true
		default:
			log.Fatalf("invalid number of arguments to return; expected 1, got %d", len(retStmt.Results))
		}
	})

	// Remove "os" import if not required.
	if !astutil.UsesImport(file, "os") {
		astutil.DeleteImport(token.NewFileSet(), file, "os")
	}

	// Remove trailing blank return statement.
	list := mainFunc.Body.List
	n := len(list)
	if n > 0 {
		if isEmptyReturn(list[n-1]) {
			mainFunc.Body.List = list[:n-1]
			fixed = true
		}
	}

	return fixed
}

// createExit creates and returns an "os.Exit" call with the specified argument.
func createExit(arg ast.Expr) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				// TODO: Locate the original identifier of "os" instead of creating
				// a new one.
				X:   ast.NewIdent("os"),
				Sel: ast.NewIdent("Exit"),
			},
			Args: []ast.Expr{
				arg,
			},
		},
	}
}

// isZero returns true if n is a integer literal with the value 0.
func isZero(n ast.Expr) bool {
	lit, ok := n.(*ast.BasicLit)
	return ok && lit.Value == "0"
}

// isEmptyReturn returns true if the given statement is an empty return
// statement (i.e. "return"), and false otherwise.
func isEmptyReturn(stmt ast.Stmt) bool {
	ret, ok := stmt.(*ast.ReturnStmt)
	return ok && len(ret.Results) == 0
}

// findMainFunc attempts to locate the "main" function of the provided file. The
// boolean value is true if successful, and false otherwise.
func findMainFunc(file *ast.File) (f *ast.FuncDecl, ok bool) {
	for _, f := range file.Decls {
		switch f := f.(type) {
		case *ast.FuncDecl:
			if f.Name.Name == "main" {
				return f, true
			}
		}
	}
	return nil, false
}
