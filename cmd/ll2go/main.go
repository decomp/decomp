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
	"sort"
	"strings"
	"unicode"

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
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	// Parse specified functions if `-funcs` is set.
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

	// Recover global variables.
	srcName := pathutil.FileName(llPath)
	file := &ast.File{}
	d := newDecompiler()
	for _, g := range module.Globals {
		global := d.globalDecl(g)
		file.Decls = append(file.Decls, global)
	}

	// Recover functions.
	var hasMain bool
	for _, f := range funcs {
		if f.Name == "main" {
			hasMain = true
		}
		var prims []*primitive.Primitive
		if len(f.Blocks) > 0 {
			prims, err = parsePrims(srcName, f.Name)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		dbg.Printf("decompiling function %q.", f.Name)
		fn, err := d.funcDecl(f, prims)
		if err != nil {
			return errors.WithStack(err)
		}
		file.Decls = append(file.Decls, fn)
	}

	// Add newIntNNN function declarations.
	var newIntSizes []int
	for newIntSize := range d.newIntSizes {
		newIntSizes = append(newIntSizes, newIntSize)
	}
	for _, newIntSize := range newIntSizes {
		x := ast.NewIdent("x")
		intType := ast.NewIdent(fmt.Sprintf("int%d", newIntSize))
		param := &ast.Field{
			Names: []*ast.Ident{x},
			Type:  intType,
		}
		retType := &ast.StarExpr{X: intType}
		result := &ast.Field{
			Type: retType,
		}
		sig := &ast.FuncType{
			Params:  &ast.FieldList{List: []*ast.Field{param}},
			Results: &ast.FieldList{List: []*ast.Field{result}},
		}
		expr := &ast.UnaryExpr{
			Op: token.AND,
			X:  x,
		}
		returnStmt := &ast.ReturnStmt{
			Results: []ast.Expr{expr},
		}
		body := &ast.BlockStmt{
			List: []ast.Stmt{returnStmt},
		}
		name := fmt.Sprintf("newInt%d", newIntSize)
		fn := &ast.FuncDecl{
			Name: ast.NewIdent(name),
			Type: sig,
			Body: body,
		}
		file.Decls = append(file.Decls, fn)
	}

	// Add types not part of builtin.
	var intSizes []int
	for intSize := range d.intSizes {
		switch intSize {
		case 8, 16, 32, 64:
			// already builtin type of Go.
		default:
			intSizes = append(intSizes, intSize)
		}
	}
	sort.Ints(intSizes)
	for _, intSize := range intSizes {
		typeName := fmt.Sprintf("int%d", intSize)
		var underlying string
		switch {
		case intSize < 8:
			underlying = "int8"
		case intSize < 16:
			underlying = "int16"
		case intSize < 32:
			underlying = "int32"
		case intSize < 64:
			underlying = "int64"
		default:
			return errors.Errorf("support for integer type with bit size %d not yet implemented", intSize)
		}
		spec := &ast.TypeSpec{
			Name: ast.NewIdent(typeName),
			Type: ast.NewIdent(underlying),
		}
		typeDecl := &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{spec},
		}
		file.Decls = append(file.Decls, typeDecl)
	}

	// Set package name.
	if hasMain {
		file.Name = ast.NewIdent("main")
	} else {
		file.Name = ident(srcName)
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
	// Global states.

	// Tracks use of integer types not part of Go builtin.
	intSizes map[int]bool
	// Tracks use of newIntNNN function calls.
	newIntSizes map[int]bool

	// Per function states.

	// Map from basic block label to conceptual basic block.
	blocks map[string]*basicBlock
	// Track use of basic block labels.
	labels map[string]bool
}

// newDecompiler returns a new decompiler.
func newDecompiler() *decompiler {
	return &decompiler{
		intSizes:    make(map[int]bool),
		newIntSizes: make(map[int]bool),
	}
}

// globalDecl converts the given LLVM IR global into a corresponding Go variable
// declaration.
func (d *decompiler) globalDecl(g *ir.Global) *ast.GenDecl {
	spec := &ast.ValueSpec{
		Names:  []*ast.Ident{d.global(g.Name)},
		Type:   d.goType(g.Typ),
		Values: []ast.Expr{d.pointerToValue(g.Init)},
	}
	return &ast.GenDecl{
		Tok:   token.VAR,
		Specs: []ast.Spec{spec},
	}
}

// pointerToValue converts the given LLVM IR value to a pointer to v and returns
// the corresponding Go expression.
func (d *decompiler) pointerToValue(v value.Value) ast.Expr {
	switch v := v.(type) {
	case *constant.Null:
		// nothing to do.
		return d.value(v)
	case *constant.Int:
		callee := fmt.Sprintf("newInt%d", v.Typ.Size)
		d.newIntSizes[v.Typ.Size] = true
		return &ast.CallExpr{
			Fun:  ast.NewIdent(callee),
			Args: []ast.Expr{d.value(v)},
		}
	case *constant.Array:
		return &ast.UnaryExpr{
			Op: token.AND,
			X:  d.value(v),
		}
	default:
		panic(fmt.Sprintf("support for value %T not yet implemented", v))
	}
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
	if len(f.Blocks) == 0 {
		return fn, nil
	}

	// Reset labels tracker.
	d.labels = make(map[string]bool)

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
				assignStmt := d.assign(phi.Name, d.value(inc.X))
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

	// A single remaining basic block indicates successful control flow recovery.
	// If more than one basic block remains, unstructured control flow is added
	// using goto-statements.
	var stmts []ast.Stmt
	for _, block := range d.blocks {
		stmts = append(stmts, d.stmts(block)...)
		stmts = append(stmts, d.term(block.Term))
	}

	// Insert labels of target branches into corresponding basic blocks.
	//for label := range d.labels {
	//	block, ok := d.blocks[label]
	//	if !ok {
	//		return nil, errors.Errorf("unable to locate basic block %q", label)
	//	}
	//	_ = block
	//	// TODO: Implement.
	//}

	body := &ast.BlockStmt{
		List: stmts,
	}
	fn.Body = body
	return fn, nil
}

// global converts the given LLVM IR global identifier to a corresponding Go
// identifier.
func (d *decompiler) global(name string) *ast.Ident {
	if isID(name) {
		name = "_" + name
	}
	return ident(name)
}

// local converts the given LLVM IR local identifier to a corresponding Go
// identifier.
func (d *decompiler) local(name string) *ast.Ident {
	if isID(name) {
		name = "_" + name
	}
	return ident(name)
}

// isID reports if the given string is an unnamed identifier.
func isID(s string) bool {
	for _, r := range s {
		if !strings.ContainsRune("0123456789", r) {
			return false
		}
	}
	return true
}

// ident returns a sanitized version of the given identifier.
func ident(s string) *ast.Ident {
	f := func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			// valid rune in identifier.
			return r
		}
		return '_'
	}
	return ast.NewIdent(strings.Map(f, s))
}

// label converts the given LLVM IR basic block label to a corresponding Go
// identifier.
func (d *decompiler) label(name string) *ast.Ident {
	name = "block_" + name
	return ident(name)
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
		return d.constant(v)
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
