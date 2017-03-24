package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	irtypes "github.com/llir/llvm/ir/types"
)

// goType converts the given LLVM IR type to a corresponding Go type.
func (d *decompiler) goType(t irtypes.Type) ast.Expr {
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
				Type: d.goType(p.Typ),
			}
			if len(p.Name) > 0 {
				param.Names = append(param.Names, d.local(p.Name))
			}
			params.List = append(params.List, param)
		}
		var results *ast.FieldList
		if !irtypes.Equal(t.Ret, irtypes.Void) {
			result := &ast.Field{
				Type: d.goType(t.Ret),
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
		d.intSizes[t.Size] = true
		return &ast.Ident{
			Name: fmt.Sprintf("int%d", t.Size),
		}
	case *irtypes.FloatType:
		panic("support for *types.FloatType not yet implemented")
	case *irtypes.PointerType:
		return &ast.StarExpr{
			X: d.goType(t.Elem),
		}
	case *irtypes.VectorType:
		panic("support for *types.VectorType not yet implemented")
	case *irtypes.LabelType:
		panic("support for *types.LabelType not yet implemented")
	case *irtypes.MetadataType:
		panic("support for *types.MetadataType not yet implemented")
	case *irtypes.ArrayType:
		return &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.FormatInt(t.Len, 10),
			},
			Elt: d.goType(t.Elem),
		}
	case *irtypes.StructType:
		panic("support for *types.StructType not yet implemented")
	default:
		panic(fmt.Sprintf("support for type %T not yet implemented", t))
	}
}
