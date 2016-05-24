package main

import (
	"go/ast"

	"github.com/mewkiz/pkg/errutil"
	"llvm.org/llvm/bindings/go/llvm"
)

// BasicBlock represents a conceptual basic block. If one statement of the basic
// block is executed all statements of the basic block are executed until the
// terminating instruction is reached which transfers control to another basic
// block.
type BasicBlock interface {
	// Name returns the name of the basic block.
	Name() string
	// Stmts returns the statements of the basic block.
	Stmts() []ast.Stmt
	// SetStmts sets the statements of the basic block.
	SetStmts(stmts []ast.Stmt)
	// Term returns the terminator instruction of the basic block.
	Term() llvm.Value
}

// basicBlock represents a basic block in which the instructions have been
// translated to Go AST statement nodes but the terminator instruction is an
// unmodified LLVM IR value.
type basicBlock struct {
	// Basic block name.
	name string
	// Basic block instructions.
	stmts []ast.Stmt
	// A map from variable name to variable definitions which represents the PHI
	// instructions of the basic block.
	phis map[string][]*definition
	// Terminator instruction.
	term llvm.Value
}

// Name returns the name of the basic block.
func (bb *basicBlock) Name() string { return bb.name }

// Stmts returns the statements of the basic block.
func (bb *basicBlock) Stmts() []ast.Stmt { return bb.stmts }

// SetStmts sets the statements of the basic block.
func (bb *basicBlock) SetStmts(stmts []ast.Stmt) { bb.stmts = stmts }

// Term returns the terminator instruction of the basic block.
func (bb *basicBlock) Term() llvm.Value { return bb.term }

// parseBasicBlock converts the provided LLVM IR basic block into a basic block
// in which the instructions have been translated to Go AST statement nodes but
// the terminator instruction is an unmodified LLVM IR value.
func parseBasicBlock(llBB llvm.BasicBlock) (bb *basicBlock, err error) {
	name, err := getBBName(llBB.AsValue())
	if err != nil {
		return nil, err
	}
	bb = &basicBlock{name: name, phis: make(map[string][]*definition)}
	for inst := llBB.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		// Handle terminator instruction.
		if inst == llBB.LastInstruction() {
			err = bb.addTerm(inst)
			if err != nil {
				return nil, errutil.Err(err)
			}
			return bb, nil
		}

		// Handle PHI instructions.
		if inst.InstructionOpcode() == llvm.PHI {
			ident, def, err := parsePHIInst(inst)
			if err != nil {
				return nil, errutil.Err(err)
			}
			bb.phis[ident] = def
			continue
		}

		// Handle non-terminator instructions.
		stmt, err := parseInst(inst)
		if err != nil {
			return nil, err
		}
		bb.stmts = append(bb.stmts, stmt)
	}
	return nil, errutil.Newf("invalid basic block %q; contains no instructions", name)
}

// addTerm adds the provided terminator instruction to the basic block. If the
// terminator instruction doesn't have a target basic block (e.g. ret) it is
// parsed and added to the statements list of the basic block instead.
func (bb *basicBlock) addTerm(term llvm.Value) error {
	// TODO: Check why there is no opcode in the llvm library for the resume
	// terminator instruction.
	switch opcode := term.InstructionOpcode(); opcode {
	case llvm.Ret:
		// The return instruction doesn't have any target basic blocks so treat it
		// like a regular instruction and append it to the list of statements.
		ret, err := parseRetInst(term)
		if err != nil {
			return err
		}
		bb.stmts = append(bb.stmts, ret)
	case llvm.Br, llvm.Switch, llvm.IndirectBr, llvm.Invoke, llvm.Unreachable:
		// Parse the terminator instruction during the control flow analysis.
		bb.term = term
	default:
		return errutil.Newf("non-terminator instruction %q at end of basic block", prettyOpcode(opcode))
	}
	return nil
}
