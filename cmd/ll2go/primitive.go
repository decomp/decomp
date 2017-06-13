package main

import (
	"fmt"
	"go/ast"
	"go/token"

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
		block.Name = prim.Entry
		block.num = condBlock.num
		return block, nil
	case "if_else":
		condName := prim.Nodes["cond"]
		condBlock, ok := d.blocks[condName]
		if !ok {
			return nil, errors.Errorf("unable to located cond basic block %q", condName)
		}
		bodyTrueName := prim.Nodes["body_true"]
		bodyTrueBlock, ok := d.blocks[bodyTrueName]
		if !ok {
			return nil, errors.Errorf("unable to located body_true basic block %q", bodyTrueName)
		}
		bodyFalseName := prim.Nodes["body_false"]
		bodyFalseBlock, ok := d.blocks[bodyFalseName]
		if !ok {
			return nil, errors.Errorf("unable to located body_false basic block %q", bodyFalseName)
		}
		exitName := prim.Nodes["exit"]
		exitBlock, ok := d.blocks[exitName]
		if !ok {
			return nil, errors.Errorf("unable to located exit basic block %q", exitName)
		}
		block, err := d.primIfElse(condBlock, bodyTrueBlock, bodyFalseBlock, exitBlock)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		block.Name = prim.Entry
		block.num = condBlock.num
		return block, nil
	case "if_return":
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
		block, err := d.primIfReturn(condBlock, bodyBlock, exitBlock)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		block.Name = prim.Entry
		block.num = condBlock.num
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
		block.Name = prim.Entry
		block.num = condBlock.num
		return block, nil
	case "post_loop":
		condName := prim.Nodes["cond"]
		condBlock, ok := d.blocks[condName]
		if !ok {
			return nil, errors.Errorf("unable to located cond basic block %q", condName)
		}
		exitName := prim.Nodes["exit"]
		exitBlock, ok := d.blocks[exitName]
		if !ok {
			return nil, errors.Errorf("unable to located exit basic block %q", exitName)
		}
		block, err := d.primPostLoop(condBlock, exitBlock)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		block.Name = prim.Entry
		block.num = condBlock.num
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
		block.Name = prim.Entry
		block.num = entryBlock.num
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

// primIfElse merges the basic blocks of the given if_else-primitive into a
// corresponding conceputal basic block for the primitive.
func (d *decompiler) primIfElse(condBlock, bodyTrueBlock, bodyFalseBlock, exitBlock *basicBlock) (*basicBlock, error) {
	// Handle terminators.
	var cond ast.Expr
	switch condTerm := condBlock.Term.(type) {
	case *ir.TermCondBr:
		cond = d.value(condTerm.Cond)
	case *ir.TermSwitch:
		cases := condTerm.Cases
		if len(cases) != 1 {
			return nil, errors.Errorf("invalid number of switch cases in if_else primitive; expected 1, got %d", len(cases))
		}
		cond = &ast.BinaryExpr{
			X:  d.value(condTerm.X),
			Op: token.EQL,
			Y:  d.constant(cases[0].X),
		}
	default:
		return nil, errors.Errorf("invalid cond terminator type; expected *ir.TermCondBr, got %T", condBlock.Term)
	}
	// TODO: Figure out a clean way to check if the body_true basic block is the
	// true branch or the false branch. If body_true is the false branch, use
	// body_true for the else body of the if-statement.
	if _, ok := bodyTrueBlock.Term.(*ir.TermBr); !ok {
		return nil, errors.Errorf("invalid body_true terminator type; expected *ir.TermBr, got %T", bodyTrueBlock.Term)
	}
	if _, ok := bodyFalseBlock.Term.(*ir.TermBr); !ok {
		return nil, errors.Errorf("invalid body_false terminator type; expected *ir.TermBr, got %T", bodyFalseBlock.Term)
	}
	block := &basicBlock{BasicBlock: &ir.BasicBlock{}}
	block.Term = exitBlock.Term
	// Handle instructions.
	block.stmts = append(block.stmts, d.stmts(condBlock)...)
	bodyTrue := &ast.BlockStmt{
		List: d.stmts(bodyTrueBlock),
	}
	bodyFalse := &ast.BlockStmt{
		List: d.stmts(bodyFalseBlock),
	}
	ifElseStmt := &ast.IfStmt{
		Cond: cond,
		Body: bodyTrue,
		Else: bodyFalse,
	}
	block.stmts = append(block.stmts, ifElseStmt)
	block.stmts = append(block.stmts, d.stmts(exitBlock)...)
	return block, nil
}

// primIfReturn merges the basic blocks of the given if_return-primitive into a
// corresponding conceputal basic block for the primitive.
func (d *decompiler) primIfReturn(condBlock, bodyBlock, exitBlock *basicBlock) (*basicBlock, error) {
	// Handle terminators.
	condTerm, ok := condBlock.Term.(*ir.TermCondBr)
	if !ok {
		return nil, errors.Errorf("invalid cond terminator type; expected *ir.TermCondBr, got %T", condBlock.Term)
	}
	cond := d.value(condTerm.Cond)
	// TODO: Figure out a clean way to check if the body basic block is the true
	// branch or the false branch. If body is the false branch, negate the
	// condition.
	bodyTermStmt := d.term(bodyBlock.Term)
	block := &basicBlock{BasicBlock: &ir.BasicBlock{}}
	block.Term = exitBlock.Term
	// Handle instructions.
	block.stmts = append(block.stmts, d.stmts(condBlock)...)
	body := &ast.BlockStmt{
		List: d.stmts(bodyBlock),
	}
	body.List = append(body.List, bodyTermStmt)
	ifReturnStmt := &ast.IfStmt{
		Cond: cond,
		Body: body,
	}
	block.stmts = append(block.stmts, ifReturnStmt)
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

// primPostLoop merges the basic blocks of the given post_loop-primitive into a
// corresponding conceputal basic block for the primitive.
func (d *decompiler) primPostLoop(condBlock, exitBlock *basicBlock) (*basicBlock, error) {
	// Handle terminators.
	condTerm, ok := condBlock.Term.(*ir.TermCondBr)
	if !ok {
		return nil, errors.Errorf("invalid cond terminator type; expected *ir.TermCondBr, got %T", condBlock.Term)
	}
	cond := d.value(condTerm.Cond)
	cond = &ast.UnaryExpr{
		Op: token.NOT,
		X:  cond,
	}
	// TODO: Figure out a clean way to check if the exit basic block is the true
	// branch or the false branch. If exit is the true branch, negate the
	// condition.
	block := &basicBlock{BasicBlock: &ir.BasicBlock{}}
	block.Term = exitBlock.Term
	// Handle instructions.
	body := &ast.BlockStmt{
		List: d.stmts(condBlock),
	}
	breakStmt := &ast.BranchStmt{Tok: token.BREAK}
	ifBreakStmt := &ast.IfStmt{
		Cond: cond,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{breakStmt},
		},
	}
	body.List = append(body.List, ifBreakStmt)
	forStmt := &ast.ForStmt{
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
