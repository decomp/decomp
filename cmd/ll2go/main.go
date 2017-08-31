// The ll2go tool decompiles LLVM IR assembly to Go source code (*.ll -> *.go).
//
// The input of ll2go is LLVM IR assembly and the output is unpolished Go source
// code.
//
// Usage:
//
//    ll2go [OPTION]... FILE.ll...
//
// Flags:
//
//    -funcs string
//          comma-separated list of functions to parse
//    -q    suppress non-error messages
package main

import (
	"bufio"
	"encoding/json"
	goerrors "errors"
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

	"github.com/decomp/decomp/cfa"
	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	irtypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/mewkiz/pkg/osutil"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
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
		file, err := ll2go(llPath, funcNames)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		// TODO: Remove debug output.
		if err := printer.Fprint(os.Stdout, token.NewFileSet(), file); err != nil {
			log.Fatalf("%+v", err)
		}
		fmt.Println()
	}
}

// ll2go converts the given LLVM IR assembly file into a corresponding Go source
// file.
func ll2go(llPath string, funcNames map[string]bool) (*ast.File, error) {
	dbg.Printf("parsing file %q.", llPath)
	module, err := asm.ParseFile(llPath)
	if err != nil {
		return nil, errors.WithStack(err)
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

	// Recover type definitions.
	srcName := pathutil.FileName(llPath)
	file := &ast.File{}
	d := newDecompiler()
	for _, t := range module.Types {
		typ := d.typeDef(t)
		file.Decls = append(file.Decls, typ)
	}

	// Recover global variables.
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
			// TODO: Clean up parsing of primitives.
			//
			//    1. Check if JSON file present on file system
			//    2. If present, parse prims from file and log to dbg that
			//       primitives are read from the JSON file.
			//    3. If not present, perform control flow analysis in memory.
			//
			// Move parts shared between restructure and ll2go to decomp/cfa.
			prims, err = parsePrims(srcName, f)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}
		dbg.Printf("decompiling function %q.", f.Name)
		fn, err := d.funcDecl(f, prims)
		if err != nil {
			return nil, errors.WithStack(err)
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
			return nil, errors.Errorf("support for integer type with bit size %d not yet implemented", intSize)
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

	return file, nil
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

// typeDef converts the given LLVM IR type into a corresponding Go type
// definition.
func (d *decompiler) typeDef(t irtypes.Type) *ast.GenDecl {
	spec := &ast.TypeSpec{
		Name: d.typeIdent(t.GetName()),
		Type: d.goTypeDef(t),
	}
	return &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{spec},
	}
}

// globalDecl converts the given LLVM IR global into a corresponding Go variable
// declaration.
func (d *decompiler) globalDecl(g *ir.Global) *ast.GenDecl {
	spec := &ast.ValueSpec{
		Names:  []*ast.Ident{d.globalIdent(g.Name)},
		Type:   d.goType(g.Typ),
		Values: []ast.Expr{d.pointerToConst(g.Init)},
	}
	return &ast.GenDecl{
		Tok:   token.VAR,
		Specs: []ast.Spec{spec},
	}
}

// pointerToConst converts the given LLVM IR constant to a pointer to c and
// returns the corresponding Go expression.
func (d *decompiler) pointerToConst(c constant.Constant) ast.Expr {
	switch c := c.(type) {
	// Simple constants
	case *constant.Int:
		callee := fmt.Sprintf("newInt%d", c.Typ.Size)
		d.newIntSizes[c.Typ.Size] = true
		return &ast.CallExpr{
			Fun:  ast.NewIdent(callee),
			Args: []ast.Expr{d.value(c)},
		}
	case *constant.Float:
		panic("support for value *constant.Float not yet implemented")
	case *constant.Null:
		// nothing to do.
		return d.value(c)
	// Complex constants
	case *constant.Vector, *constant.Array, *constant.Struct:
		return &ast.UnaryExpr{
			Op: token.AND,
			X:  d.value(c),
		}
	case *constant.ZeroInitializer:
		return &ast.CallExpr{
			Fun:  ast.NewIdent("new"),
			Args: []ast.Expr{d.goType(c.Typ)},
		}
	// Global variable and function addresses
	case *ir.Global:
		// TODO: Check if `&g` should be returned instead of `g`.
		return d.globalIdent(c.Name)
	case *ir.Function:
		// TODO: Check if `&f` should be returned instead of `f`.
		return d.globalIdent(c.Name)
	// Constant expressions
	case constant.Expr:
		panic("support for value constant.Expr not yet implemented")
	default:
		panic(fmt.Sprintf("support for value %T not yet implemented", c))
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
		Name: d.globalIdent(f.Name),
		Type: sig,
	}
	if len(f.Blocks) == 0 {
		return fn, nil
	}

	// Reset labels tracker.
	d.labels = make(map[string]bool)

	// Reset basic block mapping.
	d.blocks = make(map[string]*basicBlock)
	for i, block := range f.Blocks {
		d.blocks[block.Name] = &basicBlock{BasicBlock: block, num: i}
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
	var blocks basicBlocks
	for _, block := range d.blocks {
		blocks = append(blocks, block)
	}
	sort.Sort(blocks)
	for _, block := range blocks {
		block.stmts = d.stmts(block)
		block.stmts = append(block.stmts, d.term(block.Term))
	}

	// Insert labels of target branches into corresponding basic blocks.
	for label := range d.labels {
		// Insert label.
		block, ok := d.blocks[label]
		if !ok {
			return nil, errors.Errorf("unable to locate basic block %q", label)
		}
		if len(block.stmts) < 1 {
			// A terminator statement should always be present.
			return nil, errors.New("empty basic block; expected at least 1 statement")
		}
		labelStmt := &ast.LabeledStmt{
			Label: d.label(block.Name),
			Stmt:  block.stmts[0],
		}
		block.stmts[0] = labelStmt
	}

	var stmts []ast.Stmt
	for _, block := range blocks {
		stmts = append(stmts, block.stmts...)
	}
	body := &ast.BlockStmt{
		List: stmts,
	}
	fn.Body = body
	return fn, nil
}

// globalIdent converts the given LLVM IR type identifier to a corresponding Go
// identifier.
func (d *decompiler) typeIdent(name string) *ast.Ident {
	if isID(name) {
		name = "_" + name
	}
	return ident(name)
}

// globalIdent converts the given LLVM IR global identifier to a corresponding
// Go identifier.
func (d *decompiler) globalIdent(name string) *ast.Ident {
	if isID(name) {
		name = "_" + name
	}
	return ident(name)
}

// localIdent converts the given LLVM IR local identifier to a corresponding Go
// identifier.
func (d *decompiler) localIdent(name string) *ast.Ident {
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
	s = strings.Replace(s, ".", "dot", -1)
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
			return d.globalIdent(v.GetName())
		default:
			return d.localIdent(v.GetName())
		}
	case constant.Constant:
		return d.constant(v)
	default:
		panic(fmt.Sprintf("support for value %T not yet implemented", v))
	}
}

// intLit converts the given integer literal into a corresponding Go expression.
func (d *decompiler) intLit(i int64) ast.Expr {
	return &ast.BasicLit{
		Kind:  token.INT,
		Value: fmt.Sprintf("%d", i),
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
	// Track basic block number in f.Blocks slice, to be used for sorting basic
	// blocks after incomplete control flow recovery.
	num int
}

// basicBlocks implements the sort.Sort interface to sort basic blocks according
// to their occurrence in f.Blocks.
type basicBlocks []*basicBlock

func (bs basicBlocks) Less(i, j int) bool { return bs[i].num < bs[j].num }
func (bs basicBlocks) Len() int           { return len(bs) }
func (bs basicBlocks) Swap(i, j int)      { bs[i], bs[j] = bs[j], bs[i] }

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
func parsePrims(srcName string, f *ir.Function) ([]*primitive.Primitive, error) {
	graphsDir := fmt.Sprintf("%s_graphs", srcName)
	jsonName := f.Name + ".json"
	jsonPath := filepath.Join(graphsDir, jsonName)
	// Generate primitives if not present on file system.
	if !osutil.Exists(jsonPath) {
		prims, err := genPrims(f)
		if err != nil {
			if errors.Cause(err) == ErrIncomplete {
				dbg.Printf("WARNING: incomplete control flow recovery of %q", f.Name)
			} else {
				return nil, errors.WithStack(err)
			}
		}
		return prims, nil
	}
	// Parse primitives from file system.
	var prims []*primitive.Primitive
	fr, err := os.Open(jsonPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fr.Close()
	r := bufio.NewReader(fr)
	dec := json.NewDecoder(r)
	if err := dec.Decode(&prims); err != nil {
		return nil, errors.WithStack(err)
	}
	return prims, nil
}

// ErrIncomplete signals an incomplete control flow recovery.
var ErrIncomplete = goerrors.New("incomplete control flow recovery")

// locateEntryNode attempts to locate the entry node of the control flow graph
// by searching for a single node in the control flow graph with no incoming
// edges.
func locateEntryNode(g *cfg.Graph) (graph.Node, error) {
	var entry graph.Node
	for _, n := range g.Nodes() {
		preds := g.To(n)
		if len(preds) == 0 {
			if entry != nil {
				return nil, errors.Errorf("more than one candidate for the entry node located; prev %q, new %q", label(entry), label(n))
			}
			entry = n
		}
	}
	if entry == nil {
		return nil, errors.Errorf("unable to locate entry node; try specifying an entry node label using the -entry flag")
	}
	return entry, nil
}

// genPrims returns the high-level primitives of the given function discovered
// by control flow analysis.
func genPrims(f *ir.Function) ([]*primitive.Primitive, error) {
	g := cfg.New(f)
	entry, err := locateEntryNode(g)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	var prims []*primitive.Primitive
	for len(g.Nodes()) > 1 {
		// Locate primitive.
		dom := cfg.NewDom(g, entry)
		prim, err := cfa.FindPrim(g, dom)
		if err != nil {
			return prims, errors.Wrap(ErrIncomplete, err.Error())
		}
		prims = append(prims, prim)
		// Merge the nodes of the primitive into a single node.
		if err := cfa.Merge(g, prim); err != nil {
			return nil, errors.WithStack(err)
		}
		// Handle special case where entry node has been replaced by primitive
		// node.
		if !g.Has(entry) {
			var ok bool
			entry, ok = g.NodeByLabel(prim.Entry)
			if !ok {
				return nil, errors.Errorf("unable to locate entry node %q", prim.Entry)
			}
		}
	}
	return prims, nil
}

// label returns the label of the node.
func label(n graph.Node) string {
	if n, ok := n.(*cfg.Node); ok {
		return n.Label
	}
	panic(fmt.Sprintf("invalid node type; expected *cfg.Node, got %T", n))
}
