// TODO: Add support for parsing from standard input.

// The ll2dot tool generates control flow graphs from LLVM IR assembly (*.ll ->
// *.dot).
//
// The input of ll2dot is LLVM IR assembly and the output is a set of Graphviz
// DOT files, each representing the control flow graph of a function using one
// node per basic block.
//
// For a source file "foo.ll" containing the functions "bar" and "baz" the
// following DOT files are generated.
//
//    * foo_graphs/bar.dot
//    * foo_graphs/baz.dot
//
// Usage:
//
//    ll2dot [OPTION]... FILE.ll...
//
// Flags:
//
//    -f    force overwrite existing graph directories
//    -funcs string
//          comma-separated list of functions to parse
//    -img
//          generate an image representation of the control flow graph
//    -q    suppress non-error messages
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/decomp/decomp/graph/cfg"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
)

// dbg represents a logger with the "ll2dot:" prefix, which logs debug messages
// to standard error.
var dbg = log.New(os.Stderr, term.RedBold("ll2dot:")+" ", 0)

func usage() {
	const use = `
Generate control flow graphs from LLVM IR assembly (*.ll -> *.dot).

Usage:

	ll2dot [OPTION]... FILE.ll...

Flags:
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line flags.
	var (
		// force specifies whether to force overwrite existing graph directories.
		force bool
		// funcs represents a comma-separated list of functions to parse.
		funcs string
		// img specifies whether to generate an image representation of the
		// control flow graph.
		img bool
		// quiet specifies whether to suppress non-error messages.
		quiet bool
	)
	flag.BoolVar(&force, "f", false, "force overwrite existing graph directories")
	flag.StringVar(&funcs, "funcs", "", "comma-separated list of functions to parse")
	flag.BoolVar(&img, "img", false, "generate an image representation of the control flow graph")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	// Parse functions specified by the `-funcs` flag.
	funcNames := make(map[string]bool)
	for _, funcName := range strings.Split(funcs, ",") {
		if len(funcName) == 0 {
			continue
		}
		funcNames[funcName] = true
	}
	// Mute debug messages if `-q` is set.
	if quiet {
		dbg.SetOutput(ioutil.Discard)
	}

	// Generate control flow graphs from LLVM IR files.
	for _, llPath := range flag.Args() {
		if err := ll2dot(llPath, funcNames, force, img); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// ll2dot parses the provided LLVM IR assembly file and generates a control flow
// graph for each of its defined functions using one node per basic block.
func ll2dot(llPath string, funcNames map[string]bool, force, img bool) error {
	dbg.Printf("parsing file %q.", llPath)
	module, err := asm.ParseFile(llPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Get functions set by `-funcs` or all functions if `-funcs` not used.
	var funcs []*ir.Function
	for _, f := range module.Funcs {
		if len(funcNames) > 0 && !funcNames[f.Name] {
			dbg.Printf("skipping function %q.", f.Name)
			continue
		}
		funcs = append(funcs, f)
	}

	// Generate a control flow graph for each function.
	dotDir, err := createDotDir(llPath, force)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range funcs {
		// Skip function declarations.
		if len(f.Blocks) == 0 {
			continue
		}

		// Generate control flow graph.
		dbg.Printf("parsing function %q.", f.Name)
		g := cfg.New(f)

		// Store DOT graph.
		if err := storeCFG(g, f.Name, dotDir, img); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// createDotDir creates and returns an output directory based on the path of the
// given LLVM IR file.
//
// For a source file "foo.ll" the output directory "foo_graphs/" is created. If
// the `-force` flag is set, existing graph directories are overwritten by
// force.
func createDotDir(llPath string, force bool) (string, error) {
	dotDir := pathutil.TrimExt(llPath) + "_graphs"
	if force {
		// Force overwrite existing graph directories.
		if err := os.RemoveAll(dotDir); err != nil {
			return "", errors.WithStack(err)
		}
	}
	if err := os.Mkdir(dotDir, 0755); err != nil {
		return "", errors.WithStack(err)
	}
	return dotDir, nil
}

// storeCFG stores the given control flow graph as a DOT file. If `-img` is set,
// it also stores an image representation of the CFG.
//
// For a source file "foo.ll" containing the functions "bar" and "baz" the
// following DOT files will be created:
//
//    foo_graphs/bar.dot
//    foo_graphs/baz.dot
func storeCFG(g graph.Directed, funcName, dotDir string, img bool) error {
	buf, err := dot.Marshal(g, funcName, "", "\t", false)
	if err != nil {
		return errors.WithStack(err)
	}
	dotName := funcName + ".dot"
	dotPath := filepath.Join(dotDir, dotName)
	dbg.Printf("creating file %q.", dotPath)
	if err := ioutil.WriteFile(dotPath, buf, 0644); err != nil {
		return errors.WithStack(err)
	}
	// Store an image representation of the CFG if `-img` is set.
	if img {
		pngName := funcName + ".png"
		pngPath := filepath.Join(dotDir, pngName)
		dbg.Printf("creating file %q.", pngPath)
		cmd := exec.Command("dot", "-Tpng", "-o", pngPath, dotPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
