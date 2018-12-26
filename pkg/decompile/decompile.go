// Package decompile decompiles LLVM IR assembly to Go source code.
package decompile

import (
	"go/ast"
)

// Decompile decompiles the LLVM IR module to Go source code.
func (gen *Generator) Decompile() *ast.File {
	// Index top-level entities.
	gen.indexModule()
	// Decompile LLVM IR module to Go source code.
	gen.decompileModule()
	return gen.file
}

// decompileModule decompiles the LLVM IR module to Go source code, emitting to
// file.
func (gen *Generator) decompileModule() {
	// Index Go type definitions.
	gen.indexTypeDefs()
	// Translate LLVM IR type definitions to Go.
	gen.translateTypeDefs()

	// TODO: implement
	//panic("not yet implemented")
}
