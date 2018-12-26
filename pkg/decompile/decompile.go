// Package decompile decompiles LLVM IR assembly to Go source code.
package decompile

import (
	"go/ast"
)

// Decompile decompiles the LLVM IR module to Go source code.
func (gen *Generator) Decompile() *ast.File {
	// Resolve type definitions.

	// Index Go type definitions.
	gen.indexTypeDefs()
	// Translate LLVM IR type definitions to Go.
	gen.translateTypeDefs()

	// Index global identifiers and create scaffolding global variable and
	// function declarations.
	gen.createGlobalDecls()

	// Decompile LLVM IR module to Go source code.
	gen.decompileModule()

	return gen.file
}

// decompileModule decompiles the LLVM IR module to Go source code, emitting to
// file.
func (gen *Generator) decompileModule() {
	// Decompile LLVM IR global definitions to Go source code.
	gen.decompileGlobalDefs()

	// Decompile LLVM IR function definitions to Go source code.
	gen.decompileFuncDefs()
}

// decompileGlobalDefs decompiles the LLVM IR global definitions to Go source
// code, emitting to file.
func (gen *Generator) decompileGlobalDefs() {
	for _, irGlobal := range gen.m.Globals {
		if irGlobal.Init == nil {
			// Skip global declarations.
			continue
		}
		name := irGlobal.Name()
		global, ok := gen.globals[name]
		if !ok {
			gen.Errorf("unable to locate global variable declaration with name %q", name)
			continue
		}
		init, err := gen.liftConst(irGlobal.Init)
		if err != nil {
			gen.eh(err)
			continue
		}
		spec := global.Specs[0].(*ast.ValueSpec)
		spec.Values = []ast.Expr{init}
	}
}

// decompileFuncDefs decompiles the LLVM IR function definitions to Go source
// code, emitting to file.
func (gen *Generator) decompileFuncDefs() {
}
