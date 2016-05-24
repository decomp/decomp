//go:generate usagen ll2dot
//go:generate mv z_usage.go z_usage.bak
//go:generate mango -plain ll2dot.go
//go:generate mv z_usage.bak z_usage.go

// ll2dot is a tool which generates control flow graphs from LLVM IR assembly
// files (e.g. *.ll -> *.dot). The output is a set of Graphviz DOT files, each
// representing the control flow graph of a function using one node per basic
// block.
//
// For a source file "foo.ll" containing the functions "bar" and "baz" the
// following DOT files will be generated:
//
//    * foo_graphs/bar.dot
//    * foo_graphs/baz.dot
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

	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/pathutil"
	"llvm.org/llvm/bindings/go/llvm"
)

var (
	// When flagForce is true, force overwrite existing graph directories.
	flagForce bool
	// flagFuncs specifies a comma separated list of functions to parse (e.g.
	// "foo,bar").
	flagFuncs string
	// When flagImage is true, generate an image representation of the CFG.
	flagImage bool
	// When flagQuiet is true, suppress non-error messages.
	flagQuiet bool
)

func init() {
	flag.BoolVar(&flagForce, "f", false, "Force overwrite existing graph directories.")
	flag.StringVar(&flagFuncs, "funcs", "", `Comma separated list of functions to parse (e.g. "foo,bar").`)
	flag.BoolVar(&flagImage, "img", false, "Generate an image representation of the CFG.")
	flag.BoolVar(&flagQuiet, "q", false, "Suppress non-error messages.")
	flag.Usage = usage
}

const use = `
Usage: ll2dot [OPTION]... FILE...
Generate control flow graphs from LLVM IR assembly files (e.g. *.ll -> *.dot).

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	for _, llPath := range flag.Args() {
		err := ll2dot(llPath)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// ll2dot parses the provided LLVM IR assembly file and generates a control flow
// graph for each of its defined functions using one node per basic block.
func ll2dot(llPath string) error {
	// File name and file path without extension.
	baseName := pathutil.FileName(llPath)
	basePath := pathutil.TrimExt(llPath)

	// Create temporary foo.bc file, e.g.
	//
	//    foo.ll -> foo.bc
	bcPath := fmt.Sprintf("/tmp/%s.bc", baseName)
	cmd := exec.Command("llvm-as", "-o", bcPath, llPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errutil.Err(err)
	}

	// Remove temporary foo.bc file.
	defer func() {
		err = os.Remove(bcPath)
		if err != nil {
			log.Fatalln(errutil.Err(err))
		}
	}()

	// Create output directory for the control flow graphs.
	dotDir := basePath + "_graphs"
	if flagForce {
		// Force remove existing graph directory.
		err = os.RemoveAll(dotDir)
		if err != nil {
			return errutil.Err(err)
		}
	}
	err = os.Mkdir(dotDir, 0755)
	if err != nil {
		return errutil.Err(err)
	}

	// Parse foo.bc
	module, err := llvm.ParseBitcodeFile(bcPath)
	if err != nil {
		return errutil.Err(err)
	}
	defer module.Dispose()

	// Get function names.
	var funcNames []string
	if len(flagFuncs) > 0 {
		// Get function names from command line flag:
		//
		//    -funcs="foo,bar"
		funcNames = strings.Split(flagFuncs, ",")
	} else {
		// Get all function names.
		for f := module.FirstFunction(); !f.IsNil(); f = llvm.NextFunction(f) {
			if f.IsDeclaration() {
				// Ignore function declarations (e.g. functions without bodies).
				continue
			}
			funcNames = append(funcNames, f.Name())
		}
	}

	// Generate a control flow graph for each function.
	for _, funcName := range funcNames {
		// Generate control flow graph.
		if !flagQuiet {
			log.Printf("Parsing function: %q\n", funcName)
		}
		graph, err := createCFG(module, funcName)
		if err != nil {
			return errutil.Err(err)
		}

		// Store the control flow graph.
		//
		// For a source file "foo.ll" containing the functions "bar" and "baz" the
		// following DOT files will be created:
		//
		//    foo_graphs/bar.dot
		//    foo_graphs/baz.dot
		dotName := funcName + ".dot"
		dotPath := filepath.Join(dotDir, dotName)
		if !flagQuiet {
			log.Printf("Creating: %q\n", dotPath)
		}
		buf := []byte(graph.String())
		err = ioutil.WriteFile(dotPath, buf, 0644)
		if err != nil {
			return errutil.Err(err)
		}

		// Generate an image representation of the control flow graph.
		if flagImage {
			pngName := funcName + ".png"
			pngPath := filepath.Join(dotDir, pngName)
			if !flagQuiet {
				log.Printf("Creating: %q\n", pngPath)
			}
			cmd := exec.Command("dot", "-Tpng", "-o", pngPath, dotPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return errutil.Err(err)
			}
		}
	}

	return nil
}

// createCFG generates a control flow graph for the given function using one
// node per basic block.
func createCFG(module llvm.Module, funcName string) (*dot.Graph, error) {
	f := module.NamedFunction(funcName)
	if f.IsNil() {
		return nil, errutil.Newf("unable to locate function %q", funcName)
	}
	if f.IsDeclaration() {
		return nil, errutil.Newf("unable to generate CFG for %q; expected function definition, got function declaration (e.g. no body)", funcName)
	}

	// Create a new directed graph.
	graph := dot.NewGraph()
	graph.SetDir(true)
	graph.SetName(funcName)

	// Populate the graph with one node per basic block.
	for _, bb := range f.BasicBlocks() {
		// Add node (i.e. basic block) to the graph.
		bbName, err := getBBName(bb.AsValue())
		if err != nil {
			return nil, errutil.Err(err)
		}
		if bb == f.EntryBasicBlock() {
			attrs := map[string]string{"label": "entry"}
			graph.AddNode(funcName, bbName, attrs)
		} else {
			graph.AddNode(funcName, bbName, nil)
		}

		// Add edges from node (i.e. target basic blocks) to the graph.
		term := bb.LastInstruction()
		nops := term.OperandsCount()
		switch opcode := term.InstructionOpcode(); opcode {
		case llvm.Ret:
			// exit node.
			//    ret <type> <value>
			//    ret void

		case llvm.Br:
			switch nops {
			case 1:
				// unconditional branch.
				//    br label <target>
				target := term.Operand(0)
				targetName, err := getBBName(target)
				if err != nil {
					return nil, errutil.Err(err)
				}
				graph.AddEdge(bbName, targetName, true, nil)

			case 3:
				// 2-way conditional branch.
				//    br i1 <cond>, label <target_true>, label <target_false>
				// NOTE: The LLVM library has a peculiar way of ordering the operands:
				//    term.Operand(0) refers to the 1st operand
				//    term.Operand(1) refers to the 3rd operand
				//    term.Operand(2) refers to the 2nd operand
				// TODO: Make the order logical with the pure Go implementation of
				// LLVM IR.
				targetTrue, targetFalse := term.Operand(2), term.Operand(1)
				targetTrueName, err := getBBName(targetTrue)
				if err != nil {
					return nil, errutil.Err(err)
				}
				targetFalseName, err := getBBName(targetFalse)
				if err != nil {
					return nil, errutil.Err(err)
				}
				attrs := map[string]string{"label": "false"}
				graph.AddEdge(bbName, targetFalseName, true, attrs)
				attrs = map[string]string{"label": "true"}
				graph.AddEdge(bbName, targetTrueName, true, attrs)

			default:
				return nil, errutil.Newf("invalid number of operands (%d) for br instruction", nops)
			}

		case llvm.Switch:
			// n-way conditional branch.
			//    switch <type> <value>, label <default_target> [
			//       <type> <case1>, label <case1_target>
			//       <type> <case2>, label <case2_target>
			//       ...
			//    ]
			if nops < 2 {
				return nil, errutil.Newf("invalid number of operands (%d) for switch instruction", nops)
			}

			// Default branch.
			targetDefault := term.Operand(1)
			targetDefaultName, err := getBBName(targetDefault)
			if err != nil {
				return nil, errutil.Err(err)
			}
			attrs := map[string]string{"label": "default"}
			graph.AddEdge(bbName, targetDefaultName, true, attrs)

			// Case branches.
			for i := 3; i < nops; i += 2 {
				// Case branch.
				targetCase := term.Operand(i)
				targetCaseName, err := getBBName(targetCase)
				if err != nil {
					return nil, errutil.Err(err)
				}
				caseID := (i - 3) / 2
				label := fmt.Sprintf("case %d", caseID)
				attrs := map[string]string{"label": label}
				graph.AddEdge(bbName, targetCaseName, true, attrs)
			}

		case llvm.Unreachable:
			// unreachable node.
			//    unreachable

		default:
			// TODO: Implement support for:
			//    - indirectbr
			//    - invoke
			//    - resume
			panic(fmt.Sprintf("not yet implemented; support for terminator %v", opcode))
		}
	}

	return graph, nil
}
