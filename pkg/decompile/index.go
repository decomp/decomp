package decompile

import (
	"go/ast"
	"go/token"

	"github.com/llir/llvm/ir"
	"github.com/pkg/errors"
)

// createGlobalDecls indexes global identifiers and creates scaffolding Go
// declarations (without bodies but with types) based on the global variable and
// function declarations and definitions of the given module.
//
// post-condition: gen.globals maps from global identifier (without '@' prefix)
// to corresponding scaffolding Go global declaration.
//
// post-condition: gen.funcs maps from global identifier (without '@' prefix)
// to corresponding scaffolding Go function declaration.
func (gen *Generator) createGlobalDecls() {
	// Index global identifiers and create scaffolding global variable
	// declarations.
	for _, irGlobal := range gen.m.Globals {
		global, err := gen.newGlobal(irGlobal)
		if err != nil {
			gen.eh(err)
			continue
		}
		name := irGlobal.Name()
		if prev, ok := gen.globals[name]; ok {
			gen.Errorf("global variable declaration with name %q already present; prev `%v`, new `%v`", name, prev, global)
			continue
		}
		gen.globals[name] = global
		// Append global variable declaration to Go source file.
		gen.file.Decls = append(gen.file.Decls, global)
	}
	// Index global identifiers and create scaffolding function declarations.
	for _, irFunc := range gen.m.Funcs {
		f, err := gen.newFunc(irFunc)
		if err != nil {
			gen.eh(err)
			continue
		}
		name := irFunc.Name()
		if prev, ok := gen.funcs[name]; ok {
			gen.Errorf("function declaration with name %q already present; prev `%v`, new `%v`", name, prev, f)
			continue
		}
		gen.funcs[name] = f
		// Append function declaration to Go source file.
		gen.file.Decls = append(gen.file.Decls, f)
	}
}

// newGlobal returns a new scaffolding Go value specifier (without body but with
// type) based on the given LLVM IR global declaration or definition.
func (gen *Generator) newGlobal(irGlobal *ir.Global) (*ast.GenDecl, error) {
	name := irGlobal.Name()
	contentType, err := gen.goType(irGlobal.ContentType)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	spec := &ast.ValueSpec{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  goTypeExpr(contentType),
	}
	global := &ast.GenDecl{
		Tok:   token.VAR,
		Specs: []ast.Spec{spec},
	}
	return global, nil
}

// newFunc returns a new scaffolding Go function declaration (without body but
// with type) based on the given LLVM IR function declaration or definition.
func (gen *Generator) newFunc(irFunc *ir.Function) (*ast.FuncDecl, error) {
	panic("not yet implemented")
}
