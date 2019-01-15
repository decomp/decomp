package decompile

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/pkg/errors"
)

// liftConst lifts the LLVM IR constant to an equivalent Go expression.
func (gen *Generator) liftConst(irConst constant.Constant) (ast.Expr, error) {
	// TODO: implement support for remaining constants.
	switch irConst := irConst.(type) {
	// Simple constants
	case *constant.Int:
		return gen.liftIntConst(irConst), nil
	case *constant.Float:
		return gen.liftFloatConst(irConst), nil
	case *constant.Null:
		return ast.NewIdent("nil"), nil
	//case *constant.None:
	// Complex constants
	case *constant.Struct:
		goType, err := gen.goType(irConst.Type())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		var fields []ast.Expr
		for _, irField := range irConst.Fields {
			field, err := gen.liftConst(irField)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			fields = append(fields, field)
		}
		return &ast.CompositeLit{
			Type: goTypeExpr(goType),
			Elts: fields,
		}, nil
	case *constant.Array:
		goType, err := gen.goType(irConst.Type())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		var elems []ast.Expr
		for _, irElem := range irConst.Elems {
			elem, err := gen.liftConst(irElem)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			elems = append(elems, elem)
		}
		return &ast.CompositeLit{
			Type: goTypeExpr(goType),
			Elts: elems,
		}, nil
	case *constant.CharArray:
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: string(irConst.X),
		}, nil
	case *constant.Vector:
		goType, err := gen.goType(irConst.Type())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		var elems []ast.Expr
		for _, irElem := range irConst.Elems {
			elem, err := gen.liftConst(irElem)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			elems = append(elems, elem)
		}
		return &ast.CompositeLit{
			Type: goTypeExpr(goType),
			Elts: elems,
		}, nil
	case *constant.ZeroInitializer:
		goType, err := gen.goType(irConst.Type())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		typ := goTypeExpr(goType)
		return &ast.StarExpr{
			X: &ast.CallExpr{
				Fun:  ast.NewIdent("new"),
				Args: []ast.Expr{typ},
			},
		}, nil
	// Global variable and function addresses
	case *ir.Global:
		return ast.NewIdent(irConst.Name()), nil
	case *ir.Func:
		return ast.NewIdent(irConst.Name()), nil
	case *ir.Alias:
		return ast.NewIdent(irConst.Name()), nil
	case *ir.IFunc:
		return ast.NewIdent(irConst.Name()), nil
	// Undefined values
	case *constant.Undef:
		// TODO: figure out better representation for undefined values.
		return ast.NewIdent("nil"), nil
	// Addresses of basic blocks
	//case *constant.BlockAddress:
	// Constant expressions
	case constant.Expression:
		return gen.liftConstExpr(irConst)
	default:
		panic(fmt.Errorf("support for constant %T not yet implemented", irConst))
	}
}

// liftConstExpr lifts the LLVM IR constant expression to an equivalent Go
// expression.
func (gen *Generator) liftConstExpr(irConst constant.Constant) (ast.Expr, error) {
	switch irConst := irConst.(type) {
	// Binary expressions
	//case *constant.ExprAdd:
	//case *constant.ExprFAdd:
	//case *constant.ExprSub:
	//case *constant.ExprFSub:
	//case *constant.ExprMul:
	//case *constant.ExprFMul:
	//case *constant.ExprUDiv:
	//case *constant.ExprSDiv:
	//case *constant.ExprFDiv:
	//case *constant.ExprURem:
	//case *constant.ExprSRem:
	//case *constant.ExprFRem:
	// Bitwise expressions
	//case *constant.ExprShl:
	//case *constant.ExprLShr:
	//case *constant.ExprAShr:
	//case *constant.ExprAnd:
	//case *constant.ExprOr:
	//case *constant.ExprXor:
	// Vector expressions
	//case *constant.ExprExtractElement:
	//case *constant.ExprInsertElement:
	//case *constant.ExprShuffleVector:
	// Aggregate expressions
	//case *constant.ExprExtractValue:
	//case *constant.ExprInsertValue:
	// Memory expressions
	//case *constant.ExprGetElementPtr:
	// Conversion expressions
	//case *constant.ExprTrunc:
	//case *constant.ExprZExt:
	//case *constant.ExprSExt:
	//case *constant.ExprFPTrunc:
	//case *constant.ExprFPExt:
	//case *constant.ExprFPToUI:
	//case *constant.ExprFPToSI:
	//case *constant.ExprUIToFP:
	//case *constant.ExprSIToFP:
	//case *constant.ExprPtrToInt:
	//case *constant.ExprIntToPtr:
	//case *constant.ExprBitCast:
	//case *constant.ExprAddrSpaceCast:
	// Other expressions
	//case *constant.ExprICmp:
	//case *constant.ExprFCmp:
	//case *constant.ExprSelect:
	default:
		panic(fmt.Errorf("support for constant expression %T not yet implemented", irConst))
	}
}

// liftIntConst lifts the LLVM IR integer constant to an equivalent Go basic
// literal expression.
func (gen *Generator) liftIntConst(irConst *constant.Int) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.INT,
		Value: irConst.X.String(),
	}
}

// liftFloatConst lifts the LLVM IR floating-point constant to an equivalent Go
// basic literal expression.
func (gen *Generator) liftFloatConst(irConst *constant.Float) *ast.BasicLit {
	if irConst.NaN {
		nan := math.NaN()
		if irConst.X.Signbit() {
			nan = math.Copysign(nan, -1)
		}
		return &ast.BasicLit{
			Kind:  token.FLOAT,
			Value: strconv.FormatFloat(nan, 'g', -1, 64),
		}
	}
	return &ast.BasicLit{
		Kind:  token.FLOAT,
		Value: irConst.X.String(),
	}
}
