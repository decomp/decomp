package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(mem2varFix)
}

var mem2varFix = fix{
	name:     "mem2var",
	date:     "2019-10-30",
	f:        mem2var,
	desc:     `Promote memory to variables (_1 := new(int) -> var _1 int).`,
	disabled: false,
}

func mem2var(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    _7 := new(int32)
	//    *_7 = 0
	//    _12 = *_7
	//
	//    // to:
	//    var _7 int32
	//    _7 = 0
	//    _12 = _7
	walk(file, func(n interface{}) {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}
		if mem2varFuncDecl(funcDecl) {
			fixed = true
		}
	})

	return fixed
}

func mem2varFuncDecl(funcDecl ast.Node) bool {
	fixed := false
	walk(funcDecl, func(n interface{}) {
		stmt, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		assignStmt, ok := (*stmt).(*ast.AssignStmt)
		if !ok {
			return
		}
		if assignStmt.Tok != token.DEFINE {
			return
		}
		if len(assignStmt.Lhs) != 1 {
			return
		}
		// foo := new(int32)
		ident, ok := assignStmt.Lhs[0].(*ast.Ident)
		if !ok {
			return
		}
		callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr)
		if !ok {
			return
		}
		callee, ok := callExpr.Fun.(*ast.Ident)
		if !ok {
			return
		}
		if callee.Name != "new" {
			return
		}
		// new(int32)
		typ := callExpr.Args[0]
		// var foo int32
		*stmt = &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ident},
						Type: typ,
					},
				},
			},
		}
		fixed = true
		// rewrite memory uses to variable uses.
		//
		// from:
		//    *_16
		//
		// to:
		//    _16
		rewriteMem2Var(funcDecl, ident.Name)
	})
	return fixed
}

// rewriteMem2Var rewrites memory uses to variable uses.
//
// from:
//    *_16
//
// to:
//    _16
func rewriteMem2Var(funcDecl ast.Node, identName string) {
	walk(funcDecl, func(n interface{}) {
		expr, ok := n.(*ast.Expr)
		if !ok {
			return
		}
		// *_16
		starExpr, ok := (*expr).(*ast.StarExpr)
		if !ok {
			return
		}
		starIdent, ok := starExpr.X.(*ast.Ident)
		if !ok {
			return
		}
		if identName != starIdent.Name {
			return
		}
		*expr = ast.NewIdent(identName)
	})
}
