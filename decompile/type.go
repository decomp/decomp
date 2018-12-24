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
	case *types.FuncType:
		return gen.goFuncType(irType)
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
	case *types.StructType:
		return gen.goStructType(irType)
	default:
		panic(fmt.Errorf("support for LLVM IR type %T not yet implemented", irType))
	}
}

// goFuncType returns the Go function type corresponding to the given LLVM IR
// function type.
func (gen *Generator) goFuncType(irType *types.FuncType) (*gotypes.Signature, error) {
	// Result.
	var results *gotypes.Tuple
	if !types.Equal(irType.RetType, types.Void) {
		retType, err := gen.goType(irType.RetType)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result := gotypes.NewVar(0, nil, "", retType)
		results = gotypes.NewTuple(result)
	}
	// Parameters.
	var goParams []*gotypes.Var
	for i, irParam := range irType.Params {
		paramType, err := gen.goType(irParam)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		paramName := fmt.Sprintf("param%d", i)
		goParam := gotypes.NewVar(0, nil, paramName, paramType)
		goParams = append(goParams, goParam)
	}
	params := gotypes.NewTuple(goParams...)
	return gotypes.NewSignature(nil, params, results, irType.Variadic), nil
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

// goStructType returns the Go struct type corresponding to the given LLVM IR
// struct type.
func (gen *Generator) goStructType(irType *types.StructType) (*gotypes.Struct, error) {
	var fields []*gotypes.Var
	for i, irField := range irType.Fields {
		fieldName := fmt.Sprintf("field%d", i)
		fieldType, err := gen.goType(irField)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		field := gotypes.NewVar(0, nil, fieldName, fieldType)
		fields = append(fields, field)
	}
	return gotypes.NewStruct(fields, nil), nil
}

// --- [ Go type expressions ] -------------------------------------------------

// goTypeExpr returns the AST Go type expression corresponding to the given Go
// type.
func goTypeExpr(goType gotypes.Type) ast.Expr {
	switch goType := goType.(type) {
	case *gotypes.Array:
		return goArrayTypeExpr(goType)
	case *gotypes.Basic:
		return goBasicTypeExpr(goType)
	case *gotypes.Pointer:
		return goPointerTypeExpr(goType)
	case *gotypes.Signature:
		return goFuncTypeExpr(goType)
	case *gotypes.Struct:
		return goStructTypeExpr(goType)
	default:
		panic(fmt.Errorf("support for Go type %T not yet implemented", goType))
	}
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

// goBasicTypeExpr returns the AST Go type expression corresponding to the given
// Go basic type.
func goBasicTypeExpr(goType *gotypes.Basic) *ast.Ident {
	return ast.NewIdent(goType.Name())
}

// goFuncTypeExpr returns the AST Go type expression corresponding to the given
// Go function type.
func goFuncTypeExpr(goType *gotypes.Signature) *ast.FuncType {
	// Results.
	var results []*ast.Field
	goResults := goType.Results()
	for i := 0; i < goResults.Len(); i++ {
		goResult := goResults.At(i)
		resultType := goTypeExpr(goResult.Type())
		result := &ast.Field{
			Type: resultType,
		}
		resultName := goResult.Name()
		if len(resultName) > 0 {
			result.Names = []*ast.Ident{ast.NewIdent(resultName)}
		}
		results = append(results, result)
	}
	// Parameters.
	var params []*ast.Field
	goParams := goType.Params()
	for i := 0; i < goParams.Len(); i++ {
		goParam := goParams.At(i)
		paramType := goTypeExpr(goParam.Type())
		param := &ast.Field{
			Type: paramType,
		}
		paramName := goParam.Name()
		if len(paramName) > 0 {
			param.Names = []*ast.Ident{ast.NewIdent(paramName)}
		}
		params = append(params, param)
	}
	// TODO: figure out how to handle variadic. goType.Variadic()
	return &ast.FuncType{
		Params: &ast.FieldList{
			List: params,
		},
		Results: &ast.FieldList{
			List: results,
		},
	}
}

// goPointerTypeExpr returns the AST Go type expression corresponding to the
// given Go pointer type.
func goPointerTypeExpr(goType *gotypes.Pointer) *ast.StarExpr {
	elem := goTypeExpr(goType.Elem())
	return &ast.StarExpr{X: elem}
}

// goStructTypeExpr returns the AST Go type expression corresponding to the
// given Go struct type.
func goStructTypeExpr(goType *gotypes.Struct) *ast.StructType {
	var fields []*ast.Field
	for i := 0; i < goType.NumFields(); i++ {
		goField := goType.Field(i)
		fieldName := ast.NewIdent(goField.Name())
		fieldType := goTypeExpr(goField.Type())
		field := &ast.Field{
			Names: []*ast.Ident{fieldName},
			Type:  fieldType,
		}
		fields = append(fields, field)
	}
	return &ast.StructType{
		Fields: &ast.FieldList{
			List: fields,
		},
	}
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
