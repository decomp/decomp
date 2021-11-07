package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	irtypes "github.com/llir/llvm/ir/types"
)

// goType converts the LLVM IR type into a corresponding Go expression.
func (d *decompiler) goType(t irtypes.Type) ast.Expr {
	if name := t.Name(); len(name) > 0 {
		return d.typeIdent(name)
	}
	return d.goTypeDef(t)
}

// goTypeDef returns the definitions of the given LLVM IR type as a
// corresponding Go type.
func (d *decompiler) goTypeDef(t irtypes.Type) ast.Expr {
	switch t := t.(type) {
	case *irtypes.VoidType:
		// The void type is only valid as a function return type in LLVM IR, or as
		// part of a call instruction to a void function, or a ret instruction
		// from a void function.
		//
		// Each of these cases will be handled specifically to take consideration
		// to void types.
		panic("unexpected void type")
	case *irtypes.FuncType:
		params := &ast.FieldList{}
		for _, p := range t.Params {
			param := &ast.Field{
				Type: d.goType(p),
			}
			// TODO: figure out how to handle function parameter name.
			//paramName := p.Name()
			//if len(paramName) > 0 {
			//	param.Names = append(param.Names, d.localIdent(paramName))
			//}
			params.List = append(params.List, param)
		}
		var results *ast.FieldList
		if !irtypes.Equal(t.RetType, irtypes.Void) {
			result := &ast.Field{
				Type: d.goType(t.RetType),
			}
			results = &ast.FieldList{
				List: []*ast.Field{result},
			}
		}
		// TODO: Handle t.Variadic.
		return &ast.FuncType{
			Params:  params,
			Results: results,
		}
	case *irtypes.IntType:
		d.intSizes[t.BitSize] = true
		return &ast.Ident{
			Name: fmt.Sprintf("int%d", t.BitSize),
		}
	case *irtypes.FloatType:
		switch t.Kind {
		case irtypes.FloatKindFloat:
			return ast.NewIdent("float32")
		case irtypes.FloatKindDouble:
			return ast.NewIdent("float64")
		case irtypes.FloatKindHalf, irtypes.FloatKindFP128, irtypes.FloatKindX86_FP80, irtypes.FloatKindPPC_FP128:
			// TODO: Add proper support for non-builtin float types.
			return ast.NewIdent("float64")
		default:
			panic(fmt.Sprintf("support for floating-point kind %v not yet implemented", t.Kind))
		}
	case *irtypes.PointerType:
		return &ast.StarExpr{
			X: d.goType(t.ElemType),
		}
	case *irtypes.VectorType:
		return &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.FormatUint(t.Len, 10),
			},
			Elt: d.goType(t.ElemType),
		}
	case *irtypes.LabelType:
		panic("support for *types.LabelType not yet implemented")
	case *irtypes.MetadataType:
		panic("support for *types.MetadataType not yet implemented")
	case *irtypes.ArrayType:
		return &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.FormatUint(t.Len, 10),
			},
			Elt: d.goType(t.ElemType),
		}
	case *irtypes.StructType:
		var fs []*ast.Field
		for i, f := range t.Fields {
			name := fmt.Sprintf("field_%d", i)
			field := &ast.Field{
				Names: []*ast.Ident{d.localIdent(name)},
				Type:  d.goType(f),
			}
			fs = append(fs, field)
		}
		fields := &ast.FieldList{
			List: fs,
		}
		return &ast.StructType{
			Fields: fields,
		}
	default:
		panic(fmt.Sprintf("support for type %T not yet implemented", t))
	}
}
