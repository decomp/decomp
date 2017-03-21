package main

import (
	"fmt"
	"go/ast"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/llir/llvm/ir"
	"github.com/pkg/errors"
)

// prim merges the basic blocks of the given primitive into a corresponding
// conceputal basic block for the primitive.
func (d *decompiler) prim(prim *primitive.Primitive) (*basicBlock, error) {
	switch prim.Prim {
	case "if":
		condName := prim.Nodes["cond"]
		condBlock, ok := d.blocks[condName]
		if !ok {
			return nil, errors.Errorf("unable to located cond basic block %q", condName)
		}
		bodyName := prim.Nodes["body"]
		bodyBlock, ok := d.blocks[bodyName]
		if !ok {
			return nil, errors.Errorf("unable to located body basic block %q", bodyName)
		}
		exitName := prim.Nodes["exit"]
		exitBlock, ok := d.blocks[exitName]
		if !ok {
			return nil, errors.Errorf("unable to located exit basic block %q", exitName)
		}
		block, err := d.primIf(condBlock, bodyBlock, exitBlock)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		block.Name = prim.Node
		return block, nil
	case "pre_loop":
		condName := prim.Nodes["cond"]
		condBlock, ok := d.blocks[condName]
		if !ok {
			return nil, errors.Errorf("unable to located cond basic block %q", condName)
		}
		bodyName := prim.Nodes["body"]
		bodyBlock, ok := d.blocks[bodyName]
		if !ok {
			return nil, errors.Errorf("unable to located body basic block %q", bodyName)
		}
		exitName := prim.Nodes["exit"]
		exitBlock, ok := d.blocks[exitName]
		if !ok {
			return nil, errors.Errorf("unable to located exit basic block %q", exitName)
		}
		block, err := d.primPreLoop(condBlock, bodyBlock, exitBlock)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		block.Name = prim.Node
		return block, nil
	case "seq":
		entryName := prim.Nodes["entry"]
		entryBlock, ok := d.blocks[entryName]
		if !ok {
			return nil, errors.Errorf("unable to located entry basic block %q", entryName)
		}
		exitName := prim.Nodes["exit"]
		exitBlock, ok := d.blocks[exitName]
		if !ok {
			return nil, errors.Errorf("unable to located exit basic block %q", exitName)
		}
		block, err := d.primSeq(entryBlock, exitBlock)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		block.Name = prim.Node
		return block, nil
	default:
		panic(fmt.Sprintf("support for primitive %q not yet implemented", prim.Prim))
	}
}

// primIf merges the basic blocks of the given if-primitive into a corresponding
// conceputal basic block for the primitive.
func (d *decompiler) primIf(condBlock, bodyBlock, exitBlock *basicBlock) (*basicBlock, error) {
	// Handle terminators.
	condTerm, ok := condBlock.Term.(*ir.TermCondBr)
	if !ok {
		return nil, errors.Errorf("invalid cond terminator type; expected *ir.TermCondBr, got %T", condBlock.Term)
	}
	cond := d.value(condTerm.Cond)
	// TODO: Figure out a clean way to check if the body basic block is the true
	// branch or the false branch. If body is the false branch, negate the
	// condition.
	if _, ok := bodyBlock.Term.(*ir.TermBr); !ok {
		return nil, errors.Errorf("invalid body terminator type; expected *ir.TermBr, got %T", bodyBlock.Term)
	}
	block := &basicBlock{BasicBlock: &ir.BasicBlock{}}
	block.Term = exitBlock.Term
	// Handle instructions.
	block.stmts = append(block.stmts, d.stmts(condBlock)...)
	body := &ast.BlockStmt{
		List: d.stmts(bodyBlock),
	}
	ifStmt := &ast.IfStmt{
		Cond: cond,
		Body: body,
	}
	block.stmts = append(block.stmts, ifStmt)
	block.stmts = append(block.stmts, d.stmts(exitBlock)...)
	return block, nil
}

// primPreLoop merges the basic blocks of the given pre_loop-primitive into a
// corresponding conceputal basic block for the primitive.
func (d *decompiler) primPreLoop(condBlock, bodyBlock, exitBlock *basicBlock) (*basicBlock, error) {
	// Handle terminators.
	condTerm, ok := condBlock.Term.(*ir.TermCondBr)
	if !ok {
		return nil, errors.Errorf("invalid cond terminator type; expected *ir.TermCondBr, got %T", condBlock.Term)
	}
	cond := d.value(condTerm.Cond)
	// TODO: Figure out a clean way to check if the exit basic block is the true
	// branch or the false branch. If exit is the true branch, negate the
	// condition.
	if _, ok := bodyBlock.Term.(*ir.TermBr); !ok {
		return nil, errors.Errorf("invalid body terminator type; expected *ir.TermBr, got %T", bodyBlock.Term)
	}
	block := &basicBlock{BasicBlock: &ir.BasicBlock{}}
	block.Term = exitBlock.Term
	// Handle instructions.
	block.stmts = append(block.stmts, d.stmts(condBlock)...)
	body := &ast.BlockStmt{
		List: d.stmts(bodyBlock),
	}
	forStmt := &ast.ForStmt{
		Cond: cond,
		Body: body,
	}
	block.stmts = append(block.stmts, forStmt)
	block.stmts = append(block.stmts, d.stmts(exitBlock)...)
	return block, nil
}

// primSeq merges the basic blocks of the given seq-primitive into a
// corresponding conceputal basic block for the primitive.
func (d *decompiler) primSeq(entryBlock, exitBlock *basicBlock) (*basicBlock, error) {
	// Handle terminators.
	if _, ok := entryBlock.Term.(*ir.TermBr); !ok {
		return nil, errors.Errorf("invalid entry terminator type; expected *ir.TermBr, got %T", entryBlock.Term)
	}
	block := &basicBlock{BasicBlock: &ir.BasicBlock{}}
	block.Term = exitBlock.Term
	// Handle instructions.
	block.stmts = append(block.stmts, d.stmts(entryBlock)...)
	block.stmts = append(block.stmts, d.stmts(exitBlock)...)
	return block, nil
}
