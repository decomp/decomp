// The ll2dot tool generates control flow graphs from LLVM IR (*.ll -> *.dot).
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gonum/graph"
	"github.com/gonum/graph/encoding/dot"
	"github.com/gonum/graph/simple"
	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
)

var dbg = log.New(os.Stderr, term.RedBold("ll2dot:")+" ", 0)

func usage() {
	const use = `
ll2dot [OPTION]... FILE...
Generate control flow graphs from LLVM IR (*.ll -> *.dot).
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line flags.
	var (
		// fs represents a comma-separated list of function names.
		fs string
		// quiet specifies whether to suppress non-error messages.
		quiet bool
	)
	flag.StringVar(&fs, "funcs", "", "comma-separated list of function names")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	funcs := make(map[string]bool)
	for _, f := range strings.Split(fs, ",") {
		if len(f) < 1 {
			continue
		}
		funcs[f] = true
	}
	if quiet {
		dbg.SetOutput(ioutil.Discard)
	}

	// Parse LLVM IR files.
	for _, llPath := range flag.Args() {
		if err := ll2dot(llPath, funcs); err != nil {
			log.Fatal(err)
		}
	}
}

func ll2dot(llPath string, funcs map[string]bool) error {
	dbg.Printf("parsing file %q.", llPath)
	module, err := asm.ParseFile(llPath)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range module.Funcs {
		if len(funcs) > 0 && !funcs[f.Name] {
			// Skip function if -funcs list exist and f is not part of it.
			dbg.Printf("skipping function %q not specified in -funcs flag.", f.Name)
			continue
		}
		dbg.Printf("generating CFG for function %q.", f.Name)
		cfg := createCFG(f)
		buf, err := dot.Marshal(cfg, f.Name, "", "\t", false)
		if err != nil {
			return errors.WithStack(err)
		}
		fmt.Println(string(buf))
	}
	return nil
}

func createCFG(f *ir.Function) graph.Graph {
	cfg := simple.NewDirectedGraph(0, 0)
	nodes := make(map[string]graph.Node)
	for _, block := range f.Blocks {
		from := createNode(cfg, nodes, block.Name)
		switch term := block.Term.(type) {
		case *ir.TermRet:
			// nothing to do.
		case *ir.TermBr:
			to := createNode(cfg, nodes, term.Target.Name)
			setEdge(cfg, from, to, "")
		case *ir.TermCondBr:
			to := createNode(cfg, nodes, term.TargetTrue.Name)
			setEdge(cfg, from, to, "true")
			to = createNode(cfg, nodes, term.TargetFalse.Name)
			setEdge(cfg, from, to, "false")
		case *ir.TermSwitch:
			for _, c := range term.Cases {
				to := createNode(cfg, nodes, c.Target.Name)
				label := fmt.Sprintf("case (x=%v)", c.X.Ident())
				setEdge(cfg, from, to, label)
			}
			to := createNode(cfg, nodes, term.TargetDefault.Name)
			setEdge(cfg, from, to, "default case")
		case *ir.TermUnreachable:
			// nothing to do.
		default:
			panic(fmt.Sprintf("support for terminator %T not yet implemented", term))
		}
	}
	return cfg
}

type node struct {
	simple.Node
	attrs
}

func createNode(cfg graph.DirectedBuilder, nodes map[string]graph.Node, label string) graph.Node {
	if n, ok := nodes[label]; ok {
		return n
	}
	id := cfg.NewNodeID()
	n := &node{
		Node: simple.Node(id),
	}
	if len(label) > 0 {
		n.attrs = append(n.attrs, newLabel(label))
	}
	nodes[label] = n
	cfg.AddNode(n)
	return n
}

type edge struct {
	simple.Edge
	attrs
}

func setEdge(cfg graph.DirectedBuilder, from, to graph.Node, label string) {
	e := &edge{
		Edge: simple.Edge{
			F: from,
			T: to,
		},
	}
	if len(label) > 0 {
		e.attrs = append(e.attrs, newLabel(label))
	}
	cfg.SetEdge(e)
}

func newLabel(label string) dot.Attribute {
	return dot.Attribute{
		Key:   "label",
		Value: fmt.Sprintf("%q", label),
	}
}

type attrs []dot.Attribute

func (attrs attrs) DOTAttributes() []dot.Attribute {
	return attrs
}
