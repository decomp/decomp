package decompile

import "go/ast"

// funcGen is a Go code generator for a given function.
type funcGen struct {
	// Go source file generator.
	gen *Generator
	// Go function being generated.
	f *ast.FuncDecl
	// Current block statement being generated.
	cur *ast.BlockStmt
}

// newFuncGen returns a new Go function generator for the given Go source file
// generator.
func (gen *Generator) newFuncGen() *funcGen {
	return &funcGen{
		gen: gen,
	}
}
