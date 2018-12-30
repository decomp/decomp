// Package decompile decompiles LLVM IR assembly to Go source code.
package decompile

import (
	"go/ast"
	"log"
	"os"

	"github.com/mewkiz/pkg/term"
)

var (
	// dbg represents a logger with the "decompile:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.WhiteBold("decompile:")+" ", 0)
	// warn represents a logger with the "decompile:" prefix, which logs warning
	// messages to standard error.
	warn = log.New(os.Stderr, term.RedBold("decompile:")+" ", 0)
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
	for _, irFunc := range gen.m.Funcs {
		if len(irFunc.Blocks) == 0 {
			// Skip function declarations.
			continue
		}
		name := irFunc.Name()
		goFunc, ok := gen.funcs[name]
		if !ok {
			gen.Errorf("unable to locate function declaration with name %q", name)
			continue
		}
		fgen := gen.newFuncGen(goFunc)
		fgen.decompileFuncDef(irFunc)
	}
}
