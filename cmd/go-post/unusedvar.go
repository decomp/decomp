package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(unusedvarFix)
}

var unusedvarFix = fix{
	name:     "unusedvar",
	date:     "2019-10-31",
	f:        unusedvar,
	desc:     "Remove unused variable declarations.",
	disabled: false,
}

func unusedvar(file *ast.File) bool {
	fixed := false

	// Apply the following transitions:
	//
	// 1)
	//    // from:
	//    var x int
	//
	//    // to:
	//
	// 1)
	//    // from:
	//    var x int
	//    x = 42
	//
	//    // to:
	walk(file, func(n interface{}) {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}
		if unusedvarFuncDecl(funcDecl) {
			fixed = true
		}
	})

	return fixed
}

// unusedvarFuncDecl removes unused variable declarations from the given
// function declaration. The boolean return value indicates whether a rewrite
// was made.
func unusedvarFuncDecl(funcDecl ast.Node) bool {
	fixed := false

	unusedDecl := findUnusedDecls(funcDecl)

	// remove unused variable declarations
	walk(funcDecl, func(n interface{}) {
		s, ok := n.(*ast.Stmt)
		if !ok {
			return
		}
		switch stmt := (*s).(type) {
		case *ast.AssignStmt:
			if stmt.Tok != token.DEFINE {
				return
			}
			if len(stmt.Lhs) != 1 {
				// support for multiple variable declarations not yet implemented.
				return
			}
			ident, ok := stmt.Lhs[0].(*ast.Ident)
			if !ok {
				return
			}
			if !unusedDecl[ident.Name] {
				return
			}
			// declaration statement of unused variable declaration found.
			*s = &ast.EmptyStmt{}
			fixed = true
		case *ast.DeclStmt:
			decl, ok := stmt.Decl.(*ast.GenDecl)
			if !ok {
				return
			}
			if decl.Tok != token.VAR {
				return
			}
			if len(decl.Specs) != 1 {
				// support for multiple variable declarations not yet implemented.
				return
			}
			valueSpec, ok := decl.Specs[0].(*ast.ValueSpec)
			if !ok {
				return
			}
			if len(valueSpec.Names) != 1 {
				// support for multiple variable declarations not yet implemented.
				return
			}
			if !unusedDecl[valueSpec.Names[0].Name] {
				return
			}
			// declaration statement of unused variable declaration found.
			*s = &ast.EmptyStmt{}
			fixed = true
		}
	})

	return fixed
}

func findUnusedDecls(funcDecl ast.Node) map[string]bool {
	declNames := findDeclNames(funcDecl)
	unusedDecl := make(map[string]bool)
	// find unused variable declarations.
	walk(funcDecl, func(n interface{}) {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return
		}
		if !declNames[ident.Name] {
			return
		}
		scope := getFuncScope(funcDecl, ident)
		// unused declaration has exactly 1 use, namely the declaration statement.
		if countUses(ident, scope) != 1 {
			return
		}
		unusedDecl[ident.Name] = true
	})
	return unusedDecl
}

func findDeclNames(funcDecl ast.Node) map[string]bool {
	m := make(map[string]bool)
	walk(funcDecl, func(n interface{}) {
		switch n := n.(type) {
		case *ast.ValueSpec:
			for _, ident := range n.Names {
				m[ident.Name] = true
			}
		case *ast.AssignStmt:
			if n.Tok != token.DEFINE {
				break
			}
			for _, lhs := range n.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				m[ident.Name] = true
			}
		}
	})
	return m
}
