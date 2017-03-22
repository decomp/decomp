// TODO: Fix PHI handling; don't introduce new variables. Potentially requires
// information from data analysis.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/value"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
)

// dbg represents a logger with the "ll2go:" prefix, which logs debug messages
// to standard error.
var dbg = log.New(os.Stderr, term.GreenBold("ll2go:")+" ", 0)

func usage() {
	const use = `
Decompile LLVM IR assembly to Go source code (*.ll -> *.go).

Usage:

	ll2go [OPTION]... FILE.ll...

Flags:
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line flags.
	var (
		// funcs represents a comma-separated list of functions to parse.
		funcs string
		// quiet specifies whether to suppress non-error messages.
		quiet bool
	)
	flag.StringVar(&funcs, "funcs", "", "comma-separated list of functions to parse")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	// Parse specified functions if `-funcs` is set.
	funcNames := make(map[string]bool)
	for _, funcName := range strings.Split(funcs, ",") {
		if len(funcName) < 1 {
			continue
		}
		funcNames[funcName] = true
	}
	// Mute debug messages if `-q` is set.
	if quiet {
		dbg.SetOutput(ioutil.Discard)
	}

	// Decompile LLVM IR files to Go source code.
	for _, llPath := range flag.Args() {
		if err := ll2go(llPath, funcNames); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// ll2go converts the given LLVM IR assembly file into a corresponding Go source
// file.
func ll2go(llPath string, funcNames map[string]bool) error {
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

	srcName := pathutil.FileName(llPath)
	file := &ast.File{
		Name: ast.NewIdent(srcName),
	}
	d := newDecompiler()
	for _, f := range funcs {
		prims, err := parsePrims(srcName, f.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		dbg.Printf("decompiling function %q.", f.Name)
		fn, err := d.funcDecl(f, prims)
		if err != nil {
			return errors.WithStack(err)
		}
		file.Decls = append(file.Decls, fn)
	}
	// TODO: Remove debug output.
	if err := printer.Fprint(os.Stdout, token.NewFileSet(), file); err != nil {
		return errors.WithStack(err)
	}
	fmt.Println()
	return nil
}

// A decompiler keeps track of relevant information during the decompilation
// process.
type decompiler struct {
	// Map from basic block label to conceptual basic block.
	blocks map[string]*basicBlock
}

// newDecompiler returns a new decompiler.
func newDecompiler() *decompiler {
	return &decompiler{}
}

// funcDecl converts the given LLVM IR function into a corresponding Go function
// declaration.
func (d *decompiler) funcDecl(f *ir.Function, prims []*primitive.Primitive) (*ast.FuncDecl, error) {
	// Force generate local IDs.
	_ = f.String()

	// Recover function declaration.
	typ := d.goType(f.Sig)
	sig := typ.(*ast.FuncType)
	fn := &ast.FuncDecl{
		Name: d.global(f.Name),
		Type: sig,
	}
	if len(f.Blocks) < 1 {
		return fn, nil
	}

	// Reset basic block mapping.
	d.blocks = make(map[string]*basicBlock)
	for _, block := range f.Blocks {
		d.blocks[block.Name] = &basicBlock{BasicBlock: block}
	}

	// Record outgoing PHI values.
	for _, block := range f.Blocks {
		for _, inst := range block.Insts {
			phi, ok := inst.(*ir.InstPhi)
			if !ok {
				continue
			}
			// The incoming values of PHI instructions are propagated as assignment
			// statements to the predecessor basic blocks of the incoming values.
			for _, inc := range phi.Incs {
				pred := d.blocks[inc.Pred.Name]
				assignStmt := d.assign(d.local(phi.Name), d.value(inc.X))
				pred.out = append(pred.out, assignStmt)
			}
		}
	}

	// Recover control flow primitives.
	for _, prim := range prims {
		block, err := d.prim(prim)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// Delete merged basic blocks.
		for _, node := range prim.Nodes {
			delete(d.blocks, node)
		}
		// Add primitive basic block.
		d.blocks[block.Name] = block
	}

	// After control flow recovery, a single basic block should remain.
	var block *basicBlock
	if len(d.blocks) != 1 {
		return nil, errors.Errorf("control flow recovery failed; unable to reduce function into a single basic block; expected 1 basic block, got %d", len(d.blocks))
	}
	for _, b := range d.blocks {
		block = b
	}

	// Recover function body.
	block.stmts = append(block.stmts, d.term(block.Term))
	body := &ast.BlockStmt{
		List: block.stmts,
	}
	fn.Body = body
	return fn, nil
}

// global converts the given LLVM IR global identifier to a corresponding Go
// identifier.
func (d *decompiler) global(name string) *ast.Ident {
	return ast.NewIdent(name)
}

// local converts the given LLVM IR local identifier to a corresponding Go
// identifier.
func (d *decompiler) local(name string) *ast.Ident {
	name = "_" + name
	return ast.NewIdent(name)
}

// value converts the given LLVM IR value to a corresponding Go expression.
func (d *decompiler) value(v value.Value) ast.Expr {
	switch v := v.(type) {
	case value.Named:
		switch v.(type) {
		case *ir.Global, *ir.Function:
			return d.global(v.GetName())
		default:
			return d.local(v.GetName())
		}
	case constant.Constant:
		switch v := v.(type) {
		case *constant.Int:
			return &ast.BasicLit{
				Kind:  token.INT,
				Value: v.X.String(),
			}
		default:
			panic(fmt.Sprintf("support for constant value %T not yet implemented", v))
		}
	default:
		panic(fmt.Sprintf("support for value %T not yet implemented", v))
	}
}

// basicBlock represents a conceptual basic block, that may contain both LLVM IR
// instructions and Go statements.
type basicBlock struct {
	*ir.BasicBlock
	// Go statements.
	stmts []ast.Stmt
	// Outgoing values for PHI instructions. In other words, a list of assignment
	// statements to appear at the end of the basic block.
	out []ast.Stmt
}

// stmts converts the basic block instructions, recorded statements and outgoing
// PHI values into a corresponding list of Go statements.
func (d *decompiler) stmts(block *basicBlock) []ast.Stmt {
	var stmts []ast.Stmt
	stmts = append(stmts, d.insts(block.Insts)...)
	stmts = append(stmts, block.stmts...)
	stmts = append(stmts, block.out...)
	return stmts
}

// parsePrims parses the JSON file containing a mapping of control flow
// primitives for the given function.
func parsePrims(srcName, funcName string) ([]*primitive.Primitive, error) {
	// TODO: Generate prims if not present on file system.
	graphsDir := fmt.Sprintf("%s_graphs", srcName)
	jsonName := funcName + ".json"
	jsonPath := filepath.Join(graphsDir, jsonName)
	var prims []*primitive.Primitive
	f, err := os.Open(jsonPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	dec := json.NewDecoder(r)
	if err := dec.Decode(&prims); err != nil {
		return nil, errors.WithStack(err)
	}
	return prims, nil
}
