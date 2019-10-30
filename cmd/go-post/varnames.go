package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func init() {
	register(varnamesFix)
}

var varnamesFix = fix{
	name:     "varnames",
	date:     "2019-10-31",
	f:        varnames,
	desc:     `Assign variable names (var _1 int -> var v_1 int).`,
	disabled: false,
}

func varnames(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    var _7 int32
	//    _7 = 0
	//
	//    // to:
	//    var _7 int32
	//    _7 = 0
	walk(file, func(n interface{}) {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}
		if varnamesFuncDecl(funcDecl) {
			fixed = true
		}
	})

	return fixed
}

func varnamesFuncDecl(funcDecl ast.Node) bool {
	fixed := false
	walk(funcDecl, func(n interface{}) {
		stmt, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		declStmt, ok := (*stmt).(*ast.DeclStmt)
		if !ok {
			return
		}
		genDecl, ok := declStmt.Decl.(*ast.GenDecl)
		if !ok {
			return
		}
		valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)
		if !ok {
			return
		}
		oldIdent := valueSpec.Names[0]
		if !isLocalID(oldIdent.Name) {
			// skip variable if already named.
			return
		}
		oldName := oldIdent.Name
		newName := fmt.Sprintf("v%s", oldName)
		newIdent := ast.NewIdent(newName)
		_ = newIdent
		valueSpec.Names[0] = newIdent
		// rewrite memory uses to variable uses.
		//
		// from:
		//    *_16
		//
		// to:
		//    _16
		f := func(pos token.Pos) ast.Expr {
			return newIdent
		}
		fnot := func(pos token.Pos) ast.Expr {
			return &ast.UnaryExpr{
				Op: token.NOT,
				X:  newIdent,
			}
		}
		scope := getFuncScope(funcDecl, oldIdent)
		rewriteUses(oldIdent, f, fnot, scope)
		fixed = true
	})
	return fixed
}

func getFuncScope(funcDecl ast.Node, ident *ast.Ident) []ast.Stmt {
	var scope []ast.Stmt
	walk(funcDecl, func(n interface{}) {
		stmt, ok := n.(ast.Stmt)
		if !ok {
			return
		}
		// Only count the actual statement in which the identifier is in scope,
		// not surrounding block statements.
		if _, ok := stmt.(*ast.BlockStmt); ok {
			return
		}
		if stmtContainsIdent(stmt, ident) {
			scope = append(scope, stmt)
		}
	})
	return scope
}

func stmtContainsIdent(stmt ast.Node, ident *ast.Ident) bool {
	found := false
	walk(stmt, func(n interface{}) {
		id, ok := n.(*ast.Ident)
		if !ok {
			return
		}
		if id.Name == ident.Name {
			found = true
		}
	})
	return found
}
