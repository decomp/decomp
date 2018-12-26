package decompile

import (
	"go/ast"
	gotypes "go/types"

	"github.com/llir/llvm/ir"
)

// Generator keeps track of top-level declarations when decompiling from LLVM IR
// to Go AST representation.
type Generator struct {
	// Error handler used to report errors encountered during decompilation.
	eh func(error)
	// LLVM IR module being decompiled.
	m *ir.Module
	// Go source file being generated.
	file *ast.File

	// Index of Go top-level declarations.

	// typeDefs maps from type name to type definition.
	typeDefs map[string]*gotypes.Named
	// globals maps from global name to global declarations and defintions.
	globals map[string]*ast.ValueSpec
	// funcs maps from global name to function declarations and defintions.
	funcs map[string]*ast.FuncDecl
}

// NewGenerator returns a new generator for decompiling the LLVM IR module to Go
// source code. The error handler eh is invoked when an error is encountered
// during decompilation.
func NewGenerator(eh func(error), m *ir.Module) *Generator {
	// TODO: sanitize source file name to be valid Go identifier.
	pkgName := m.SourceFilename
	if len(pkgName) == 0 {
		pkgName = "p"
	}
	for _, f := range m.Funcs {
		if f.Name() == "main" {
			pkgName = "main"
			break
		}
	}
	gen := &Generator{
		eh: eh,
		m:  m,
		file: &ast.File{
			Name: ast.NewIdent(pkgName),
		},
		typeDefs: make(map[string]*gotypes.Named),
		globals:  make(map[string]*ast.ValueSpec),
		funcs:    make(map[string]*ast.FuncDecl),
	}
	return gen
}
