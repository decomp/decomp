package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gonum/graph"
	"github.com/gonum/graph/encoding/dot"
	"github.com/gonum/graph/simple"
	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"
	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	for _, llPath := range flag.Args() {
		if err := ll2dot(llPath); err != nil {
			log.Fatal(err)
		}
	}
}

func ll2dot(llPath string) error {
	module, err := asm.ParseFile(llPath)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range module.Funcs {
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
	labelAttr
}

func createNode(cfg graph.DirectedBuilder, nodes map[string]graph.Node, label string) graph.Node {
	if n, ok := nodes[label]; ok {
		return n
	}
	id := cfg.NewNodeID()
	n := &node{
		Node:      simple.Node(id),
		labelAttr: labelAttr(label),
	}
	nodes[label] = n
	cfg.AddNode(n)
	return n
}

type edge struct {
	simple.Edge
	labelAttr
}

func setEdge(cfg graph.DirectedBuilder, from, to graph.Node, label string) {
	e := &edge{
		Edge: simple.Edge{
			F: from,
			T: to,
		},
		labelAttr: labelAttr(label),
	}
	cfg.SetEdge(e)
}

type labelAttr string

func (label labelAttr) DOTAttributes() []dot.Attribute {
	var attrs []dot.Attribute
	if len(label) > 0 {
		attr := dot.Attribute{
			Key:   "label",
			Value: fmt.Sprintf("%q", label),
		}
		attrs = append(attrs, attr)
	}
	return attrs
}
