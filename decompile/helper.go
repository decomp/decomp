package decompile

import (
	"go/ast"
	"go/token"
	"strconv"
)

// goIntLit returns the AST Go integer literal corresponding to the given 64-bit
// integer.
func goIntLit(n int64) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.INT,
		Value: strconv.FormatInt(n, 10),
	}
}
