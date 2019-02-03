// The ll2go tool decompiles LLVM IR assembly to Go source code (*.ll -> *.go).
//
// The input of ll2go is LLVM IR assembly and the output is unpolished Go source
// code.
//
// Usage:
//
//     ll2go [OPTION]... [FILE.ll]
//
// Flags:
//
//   -funcs string
//         comma-separated list of functions to parse
//   -o string
//         output path
//   -q    suppress non-error messages
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/mewkiz/pkg/jsonutil"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/mewmew/lnp/pkg/cfa/primitive"
	"github.com/mewmew/lnp/pkg/decompile"
	"github.com/pkg/errors"
)

var (
	// dbg represents a logger with the "ll2go:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.CyanBold("ll2go:")+" ", 0)
	// warn represents a logger with the "ll2go:" prefix, which logs warning
	// messages to standard error.
	warn = log.New(os.Stderr, term.RedBold("ll2go:")+" ", 0)
)

func usage() {
	const use = `
Decompile LLVM IR assembly to Go source code (*.ll -> *.go).

Usage:

	ll2go [OPTION]... [FILE.ll]

Flags:
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line arguments.
	var (
		// funcs represents a comma-separated list of functions to parse.
		funcs string
		// output specifies the output path.
		output string
		// quiet specifies whether to suppress non-error messages.
		quiet bool
	)
	flag.StringVar(&funcs, "funcs", "", "comma-separated list of functions to parse")
	flag.StringVar(&output, "o", "", "output path")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.Usage = usage
	flag.Parse()
	var llPath string
	switch flag.NArg() {
	case 0:
		// Parse LLVM IR assembly file from standard input.
		llPath = "-"
	case 1:
		llPath = flag.Arg(0)
	default:
		flag.Usage()
		os.Exit(1)
	}
	// Parse functions specified by the `-funcs` flag.
	funcNames := make(map[string]bool)
	for _, funcName := range strings.Split(funcs, ",") {
		funcName = strings.TrimSpace(funcName)
		if len(funcName) == 0 {
			continue
		}
		funcNames[funcName] = true
	}
	if quiet {
		// Mute debug messages if `-q` is set.
		dbg.SetOutput(ioutil.Discard)
	}

	// Parse LLMV IR assembly file.
	m, err := parseModule(llPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// Decompile LLVM IR assembly to Go source code.
	file, err := ll2go(m, llPath, funcNames)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// Output Go source file.
	w := os.Stdout
	if len(output) > 0 {
		f, err := os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w = f
	}
	if err := outputGo(w, file); err != nil {
		log.Fatalf("%+v", err)
	}
}

// ll2go decompiles the given LLVM IR module into a corresponding Go source
// file.
//
// llPath specifies the path to the LLVM IR assembly file, relative to which the
// output of other analysis phases are located.
//
// funcNames specifies the set of function names to decompile. When funcNames is
// emtpy, all functions of the module are decompiled.
func ll2go(m *ir.Module, llPath string, funcNames map[string]bool) (*ast.File, error) {
	// Error handler.
	var errs ErrorList
	eh := func(err error) {
		errs = append(errs, err)
	}
	// Decompile LLVM IR module to Go source code.
	gen := decompile.NewGenerator(eh, m)
	// Set function for parsing recovered control flow primitives.
	gen.Prims = func(f *ir.Func) ([]*primitive.Primitive, error) {
		return parsePrims(llPath, f.Name())
	}
	file := gen.Decompile()
	if len(errs) > 0 {
		// TODO: return partial results of decompilation?
		return nil, errs
	}
	return file, nil
}

// parseModule parses the given LLVM IR assembly file into an LLVM IR module.
func parseModule(llPath string) (*ir.Module, error) {
	switch llPath {
	case "-":
		// Parse LLVM IR module from standard input.
		dbg.Printf("parsing standard input")
		return asm.Parse("stdin", os.Stdin)
	default:
		dbg.Printf("parsing file %q", llPath)
		return asm.ParseFile(llPath)
	}
}

// parsePrims parses the recovered control flow primitives of the given
// function.
func parsePrims(llPath, funcName string) ([]*primitive.Primitive, error) {
	dotDir := pathutil.TrimExt(llPath) + "_graphs"
	jsonName := funcName + ".json"
	jsonPath := filepath.Join(dotDir, jsonName)
	var prims []*primitive.Primitive
	if err := jsonutil.ParseFile(jsonPath, &prims); err != nil {
		return nil, errors.WithStack(err)
	}
	return prims, nil
}

// outputGo outputs the given Go source file, writing to w.
func outputGo(w io.Writer, file *ast.File) error {
	if err := printer.Fprint(w, token.NewFileSet(), file); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ### [ Helper functions ] ####################################################

// ErrorList is a list of zero or more errors.
type ErrorList []error

// Error implements the error interface for ErrorList.
func (errs ErrorList) Error() string {
	switch len(errs) {
	case 0:
		panic("invalid call to ErrorList.Error; error list is empty")
	case 1:
		return fmt.Sprintf("error during compilation: %v", errs[0])
	default:
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "%d errors during compilation:", len(errs))
		for _, err := range errs {
			fmt.Fprintf(buf, "\n\t%v", err)
		}
		return buf.String()
	}
}
