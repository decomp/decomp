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
//     ll2dot [OPTION]... [FILE.ll]...
//
// Flags:
//
//   -f    force overwrite existing graph directories
//   -funcs string
//         comma-separated list of functions to parse
//   -img
//         output image representation of graphs
//   -q    suppress non-error messages
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

	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/mewmew/lnp/pkg/cfg"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

var (
	// dbg represents a logger with the "ll2dot:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.GreenBold("ll2dot:")+" ", 0)
	// warn represents a logger with the "ll2dot:" prefix, which logs warning
	// messages to standard error.
	warn = log.New(os.Stderr, term.RedBold("ll2dot:")+" ", 0)
)

func usage() {
	const use = `
Generate control flow graphs from LLVM IR assembly (*.ll -> *.dot).

Usage:

	ll2dot [OPTION]... [FILE.ll]...

Flags:
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line arguments.
	var (
		// force specifies whether to force overwrite existing graph directories.
		force bool
		// funcs represents a comma-separated list of functions to parse.
		funcs string
		// img specifies whether to output image representation of graphs.
		img bool
		// quiet specifies whether to suppress non-error messages.
		quiet bool
	)
	flag.BoolVar(&force, "f", false, "force overwrite existing graph directories")
	flag.StringVar(&funcs, "funcs", "", "comma-separated list of functions to parse")
	flag.BoolVar(&img, "img", false, "output image representation of graphs")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.Usage = usage
	flag.Parse()
	var llPaths []string
	switch flag.NArg() {
	case 0:
		// Parse LLVM IR module from standard input.
		llPaths = []string{"-"}
	default:
		llPaths = flag.Args()
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

	// Generate control flow graphs from LLVM IR files.
	for _, llPath := range llPaths {
		// Parse LLMV IR assembly file.
		m, err := parseModule(llPath)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		if len(m.Funcs) == 0 {
			warn.Printf("no functions in module %q", llPath)
			continue
		}
		// Create output directory.
		dotDir, err := createDotDir(llPath, force)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		// Generate control flow graphs.
		if err := ll2dot(m, dotDir, funcNames, force, img); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// ll2dot generates a control flow graph for each of the functions defined in
// the given LLVM IR module, using one node per basic block.
//
// dotDir specifies the output directory for the generated control flow graphs.
// When force is set, the output directory is automatically cleaned before
// generating new graphs.
//
// funcNames specifies the set of function names for which to generate control
// flow graphs. When funcNames is emtpy, control flow graphs are generated for
// all function definitions of the module.
//
// img specifies whether to output image representations of the control flow
// graphs.
func ll2dot(m *ir.Module, dotDir string, funcNames map[string]bool, force, img bool) error {
	// Get functions set by `-funcs` or all functions if `-funcs` not used.
	var funcs []*ir.Function
	for _, f := range m.Funcs {
		if len(funcNames) > 0 && !funcNames[f.Name()] {
			dbg.Printf("skipping function %q.", f.Ident())
			continue
		}
		funcs = append(funcs, f)
	}

	// Generate a control flow graph for each function.
	for _, f := range funcs {
		// Skip function declarations.
		if len(f.Blocks) == 0 {
			continue
		}
		// Generate control flow graph.
		dbg.Printf("parsing function %q.", f.Ident())
		g := cfg.NewGraphFromFunc(f)

		// Output control flow graph in Graphviz DOT format.
		if err := outputCFG(g, f.Name(), dotDir, img); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// createDotDir creates and returns an output directory based on the path of the
// LLVM IR file.
//
// For a source file "foo.ll" the output directory "foo_graphs/" is created. If
// the `-force` flag is set, existing graph directories are overwritten by
// force.
func createDotDir(llPath string, force bool) (string, error) {
	var dotDir string
	switch llPath {
	case "-":
		dotDir = "stdin_graphs"
	default:
		dotDir = pathutil.TrimExt(llPath) + "_graphs"
	}
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

// parseModule parses the given LLVM IR assembly file into an LLVM IR module.
func parseModule(llPath string) (*ir.Module, error) {
	switch llPath {
	case "-":
		// Parse LLVM IR module from standard input.
		dbg.Printf("parsing standard input.")
		return asm.Parse("stdin", os.Stdin)
	default:
		dbg.Printf("parsing file %q.", llPath)
		return asm.ParseFile(llPath)
	}
}

// outputCFG outputs the given control flow graph in Graphviz DOT format. If img
// is set, it also stores an image representation of the control flow graph.
//
// For a source file "foo.ll" containing the functions "bar" and "baz" the
// following DOT files will be created:
//
//    foo_graphs/bar.dot
//    foo_graphs/baz.dot
func outputCFG(g graph.Directed, funcName, dotDir string, img bool) error {
	buf, err := dot.Marshal(g, fmt.Sprintf("%q", funcName), "", "\t")
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
