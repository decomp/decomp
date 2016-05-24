// HACK: This entire file is a hack!
//
// LLVM IR has a notion of unnamed variables and basic blocks which are given
// function scoped IDs during assembly generation. The in-memory representation
// does not include this ID, so instead of reimplementing the logic of ID slots
// we capture the output of Value.Dump to locate the basic block names. Note
// that unnamed basic blocks are not given explicit labels during vanilla LLVM
// assembly generation, but rather comments which include the basic block ID.
// For this reason the "unnamed.patch" has been applied to the LLVM code base,
// which ensures that all basic blocks are given explicit labels.

package main

// #include <stdio.h>
//
// void fflush_stderr(void) {
// 	fflush(stderr);
// }
import "C"

import (
	"io/ioutil"

	"github.com/llir/llvm/asm/lexer"
	"github.com/llir/llvm/asm/token"
	"github.com/mewkiz/pkg/errutil"
	"golang.org/x/sys/unix"
	"llvm.org/llvm/bindings/go/llvm"
)

// getBBName returns the name (or ID if unnamed) of a basic block.
func getBBName(v llvm.Value) (string, error) {
	if !v.IsBasicBlock() {
		return "", errutil.Newf("invalid value type; expected basic block, got %v", v.Type())
	}

	// Locate the name of a named basic block.
	if name := v.Name(); len(name) > 0 {
		return name, nil
	}

	// Locate the ID of an unnamed basic block by parsing the value dump in
	// search for its basic block label.
	//
	// Example value dump:
	//    0:
	//      br i1 true, label %1, label %2
	//
	// Each basic block is expected to have a label, which requires the
	// "unnamed.patch" to be applied to the llvm.org/llvm/bindings/go/llvm code
	// base.
	s, err := hackDump(v)
	if err != nil {
		return "", errutil.Err(err)
	}
	tokens := lexer.ParseString(s)
	if len(tokens) < 1 {
		return "", errutil.Newf("unable to locate basic block label in %q", s)
	}
	tok := tokens[0]
	if tok.Kind != token.Label {
		return "", errutil.Newf("invalid token; expected %v, got %v", token.Label, tok.Kind)
	}
	return tok.Val, nil
}

// hackDump returns the value dump as a string.
func hackDump(v llvm.Value) (string, error) {
	// Open temp file.
	// TODO: Use an in-memory file instead of /tmp/x.
	fd, err := unix.Open("/tmp/x", unix.O_WRONLY|unix.O_TRUNC|unix.O_CREAT, 0644)
	if err != nil {
		return "", errutil.Err(err)
	}

	// Store original stderr.
	stderr, err := unix.Dup(2)
	if err != nil {
		return "", errutil.Err(err)
	}

	// Capture stderr and redirect its output to the temp file.
	err = unix.Dup2(fd, 2)
	if err != nil {
		return "", errutil.Err(err)
	}
	err = unix.Close(fd)
	if err != nil {
		return "", errutil.Err(err)
	}

	// Dump value.
	v.Dump()
	C.fflush_stderr()

	// Restore stderr.
	err = unix.Dup2(stderr, 2)
	if err != nil {
		return "", errutil.Err(err)
	}
	err = unix.Close(stderr)
	if err != nil {
		return "", errutil.Err(err)
	}

	// Return content of temp file.
	buf, err := ioutil.ReadFile("/tmp/x")
	if err != nil {
		return "", errutil.Err(err)
	}
	return string(buf), nil
}
