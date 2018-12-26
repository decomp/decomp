package decompile

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
	"strconv"

	"github.com/llir/llvm/ir/constant"
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
	//case *constant.Null:
	//case *constant.None:
	// Complex constants
	//case *constant.Struct:
	//case *constant.Array:
	//case *constant.CharArray:
	//case *constant.Vector:
	//case *constant.ZeroInitializer:
	// Global variable and function addresses
	//case *ir.Global:
	//case *ir.Function:
	//case *ir.Alias:
	//case *ir.IFunc:
	// Undefined values
	//case *constant.Undef:
	// Addresses of basic blocks
	//case *constant.BlockAddress:
	// Constant expressions
	//case constant.Expression:
	default:
		panic(fmt.Errorf("support for constant %T not yet implemented", irConst))
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
