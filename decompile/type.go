package decompile

import (
	"fmt"
	gotypes "go/types"

	"github.com/llir/llvm/ir/types"
)

// goType returns the Go type corresponding to the given LLVM IR type.
func (gen *Generator) goType(irType types.Type) (gotypes.Type, error) {
	switch irType := irType.(type) {
	default:
		panic(fmt.Errorf("support for LLVM IR type %T not yet implemented", irType))
	}
}
