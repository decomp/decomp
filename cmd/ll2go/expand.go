package main

import (
	"go/ast"

	"github.com/mewkiz/pkg/errutil"
)

// expand attempts to locate and return the definition of the provided
// identifier expression. The assignment statement defining the identifier is
// removed from the statement list of the basic block.
func expand(bb BasicBlock, ident ast.Expr) (ast.Expr, error) {
	id, ok := ident.(*ast.Ident)
	if !ok {
		return nil, errutil.Newf("unable to expand expression; expected identifier, got %T", ident)
	}
	var stmts []ast.Stmt
	var expr ast.Expr
	for _, stmt := range bb.Stmts() {
		switch stmt := stmt.(type) {
		case *ast.AssignStmt:
			if sameIdent(stmt.Lhs, id) {
				if len(stmt.Rhs) != 1 {
					return nil, errutil.Newf("invalid right-hand side; expected length 1, got %d", len(stmt.Rhs))
				}
				expr = stmt.Rhs[0]
				// TODO: Verify if the identifier is used by any other statement or
				// terminator instruction before removing it; this includes the use
				// within other basic blocks.

				// Remove the variable definition of the right-hand side expression
				// by not appending the assignment statement to the statement list.
				continue
			}
		}
		stmts = append(stmts, stmt)
	}
	if expr == nil {
		return nil, errutil.Newf("unable to expand definition of identifier %q", id.Name)
	}
	bb.SetStmts(stmts)
	return expr, nil
}

// sameIdent returns true if the left-hand side expression list contains only
// the given expression, and false otherwise.
func sameIdent(lhs []ast.Expr, ident *ast.Ident) bool {
	if len(lhs) != 1 {
		return false
	}
	lident, ok := lhs[0].(*ast.Ident)
	if !ok {
		return false
	}
	return lident.Name == ident.Name
}
