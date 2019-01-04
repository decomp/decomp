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
	true,
}

// localid replaces the use of local variable IDs with their definition. The
// boolean return value indicates that the AST was updated.
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
			// Check use count of ident. Only rewrite if used exactly once.
			scope := getScope(file, ident)
			totalUses, lhsUses := countUses(ident, scope), countUsesLhs(ident, scope)
			//fmt.Printf("ident %q used %d times, %d on lhs\n", name, totalUses, lhsUses)
			if !(totalUses == 2 && lhsUses == 1) {
				return
			}

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
		// Only count the actual statement in which the identifier is in scope,
		// not surrounding block statements.
		switch stmt := stmt.(type) {
		// TODO: add remaining statements besides BlockStmt.
		case *ast.AssignStmt, *ast.ExprStmt, *ast.ReturnStmt:
			if containsIdent(stmt, ident) {
				scope = append(scope, stmt)
			}
		case *ast.ForStmt:
			if containsIdent(stmt.Cond, ident) {
				scope = append(scope, stmt)
			}
		case *ast.IfStmt:
			if containsIdent(stmt.Cond, ident) {
				scope = append(scope, stmt)
			}
		}
	}
	walk(file, f)
	return scope
}

// containsIdent reports if the node contains the given identifier.
func containsIdent(n ast.Node, ident *ast.Ident) bool {
	found := false
	f := func(n interface{}) {
		expr, ok := n.(ast.Expr)
		if !ok {
			return
		}
		if refersTo(expr, ident) {
			found = true
		}
	}
	walk(n, f)
	return found
}

// countUsesLhs returns the number of uses on the left-hand side of the
// identifier x in scope.
func countUsesLhs(x *ast.Ident, scope []ast.Stmt) int {
	count := 0
	ff := func(n interface{}) {
		if n, ok := n.(*ast.AssignStmt); ok {
			for _, expr := range n.Lhs {
				if containsIdent(expr, x) {
					count++
				}
			}
		}
	}
	for _, n := range scope {
		walk(n, ff)
	}
	return count
}

// isLocalID reports whether the given variable name is a local ID (e.g. "_42").
func isLocalID(name string) bool {
	if strings.HasPrefix(name, "_") {
		if _, err := strconv.Atoi(name[1:]); err == nil {
			return true
		}
	}
	return false
}
