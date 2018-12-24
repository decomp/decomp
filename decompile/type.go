package decompile

import (
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"

	"github.com/llir/llvm/ir/types"
	"github.com/pkg/errors"
)

// translateTypeDefs translates the type definitions of the LLVM IR module to
// equivalent Go type definitions.
func (gen *Generator) translateTypeDefs() {
	for _, irTypeDef := range gen.m.TypeDefs {
		typeName := irTypeDef.Name()
		goType, err := gen.goType(irTypeDef)
		if err != nil {
			gen.eh(err)
			continue
		}
		// Append Go type definition to Go source file.
		typeDecl := newTypeDef(typeName, goType)
		gen.file.Decls = append(gen.file.Decls, typeDecl)
	}
}

// goType returns the Go type corresponding to the given LLVM IR type.
func (gen *Generator) goType(irType types.Type) (gotypes.Type, error) {
	switch irType := irType.(type) {
	case *types.VoidType:
		// Void types are not present in the Go type system. When translating
		// types using void types (e.g. function types), they should simply be
		// ignored directly.
		panic("cannot represent LLVM IR void type as Go type")
	//case *types.FuncType:
	case *types.IntType:
		return gen.goIntType(irType), nil
	case *types.FloatType:
		return gen.goFloatType(irType), nil
	//case *types.MMXType:
	case *types.PointerType:
		return gen.goPointerType(irType)
	//case *types.VectorType:
	//case *types.LabelType:
	//case *types.TokenType:
	//case *types.MetadataType:
	case *types.ArrayType:
		return gen.goArrayType(irType)
	//case *types.StructType:
	default:
		panic(fmt.Errorf("support for LLVM IR type %T not yet implemented", irType))
	}
}

// goIntType returns the Go integer type corresponding to the given LLVM IR
// integer type.
func (gen *Generator) goIntType(irType *types.IntType) *gotypes.Basic {
	// TODO: figure out how to distinguish signed vs. unsigned integer types.
	switch irType.BitSize {
	case 1:
		return gotypes.Typ[gotypes.Bool]
	case 8:
		return gotypes.Typ[gotypes.Int8]
	case 16:
		return gotypes.Typ[gotypes.Int16]
	case 32:
		return gotypes.Typ[gotypes.Int32]
	case 64:
		return gotypes.Typ[gotypes.Int64]
	default:
		panic(fmt.Errorf("support for integer type bit size %d not yet implemented", irType.BitSize))
	}
}

// goFloatType returns the Go floating-point type corresponding to the given
// LLVM IR floating-point type.
func (gen *Generator) goFloatType(irType *types.FloatType) *gotypes.Basic {
	switch irType.Kind {
	//case types.FloatKindHalf:
	case types.FloatKindFloat:
		return gotypes.Typ[gotypes.Float32]
	case types.FloatKindDouble:
		return gotypes.Typ[gotypes.Float64]
	//case types.FloatKindFP128:
	//case types.FloatKindX86_FP80:
	//case types.FloatKindPPC_FP128:
	default:
		panic(fmt.Errorf("support for floating-point type kind %v not yet implemented", irType.Kind))
	}
}

// goPointerType returns the Go pointer type corresponding to the given LLVM IR
// pointer type.
func (gen *Generator) goPointerType(irType *types.PointerType) (*gotypes.Pointer, error) {
	elem, err := gen.goType(irType.ElemType)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return gotypes.NewPointer(elem), nil
}

// goArrayType returns the Go array type corresponding to the given LLVM IR
// array type.
func (gen *Generator) goArrayType(irType *types.ArrayType) (*gotypes.Array, error) {
	elem, err := gen.goType(irType.ElemType)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return gotypes.NewArray(elem, int64(irType.Len)), nil
}

// ### [ Helper functions ] ####################################################

// newTypeDef returns a new Go type definition based on the given type name and
// Go type.
func newTypeDef(name string, goType gotypes.Type) *ast.GenDecl {
	spec := &ast.TypeSpec{
		Name: ast.NewIdent(name),
		Type: goTypeExpr(goType),
	}
	return &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{spec},
	}
}

// goTypeExpr returns the AST Go type expression corresponding to the given Go
// type.
func goTypeExpr(goType gotypes.Type) ast.Expr {
	switch goType := goType.(type) {
	case *gotypes.Basic:
		return goBasicTypeExpr(goType)
	case *gotypes.Pointer:
		return goPointerTypeExpr(goType)
	case *gotypes.Array:
		return goArrayTypeExpr(goType)
	default:
		panic(fmt.Errorf("support for Go type %T not yet implemented", goType))
	}
}

// goBasicTypeExpr returns the AST Go type expression corresponding to the given
// Go basic type.
func goBasicTypeExpr(goType *gotypes.Basic) *ast.Ident {
	return ast.NewIdent(goType.Name())
}

// goPointerTypeExpr returns the AST Go type expression corresponding to the
// given Go pointer type.
func goPointerTypeExpr(goType *gotypes.Pointer) *ast.StarExpr {
	elem := goTypeExpr(goType.Elem())
	return &ast.StarExpr{X: elem}
}

// goArrayTypeExpr returns the AST Go type expression corresponding to the given
// Go array type.
func goArrayTypeExpr(goType *gotypes.Array) *ast.ArrayType {
	elem := goTypeExpr(goType.Elem())
	return &ast.ArrayType{
		Len: goIntLit(goType.Len()),
		Elt: elem,
	}
}
