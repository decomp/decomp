package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/mewmew/lnp/cfa"
	"github.com/mewmew/lnp/cfa/hammock"
	"github.com/mewmew/lnp/cfa/primitive"
	"github.com/mewmew/lnp/cfg"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/encoding"
)

var (
	// dbg represents a logger with the "restructure:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.MagentaBold("restructure:")+" ", 0)
	// warn represents a logger with the "restructure:" prefix, which logs
	// warning messages to standard error.
	warn = log.New(os.Stderr, term.RedBold("restructure:")+" ", 0)
)

func main() {
	// Parse command line arguments.
	var (
		// steps specifies whether to output intermediate steps.
		steps bool
	)
	flag.BoolVar(&steps, "steps", false, "output intermediate steps")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	dotPath := flag.Arg(0)

	// Recover high-level control flow primitives.
	if err := restructure(dotPath, steps); err != nil {
		log.Fatalf("%+v", err)
	}
}

func restructure(dotPath string, steps bool) error {
	// Parse control flow graph.
	g, err := cfg.ParseFile(dotPath)
	if err != nil {
		return errors.WithStack(err)
	}
	// Output intermediate steps in Graphviz DOT format.
	var (
		before func(g cfa.Graph, prim *primitive.Primitive)
		after  func(g cfa.Graph, prim *primitive.Primitive)
	)
	step := 1
	basePath := pathutil.TrimExt(dotPath)
	if steps {
		before = func(g cfa.Graph, prim *primitive.Primitive) {
			data := []byte(dotBeforeMerge(g, prim))
			dbg.Printf("located primitive:\n%s\n", prim)
			beforePath := fmt.Sprintf("%s_xx_%04da.dot", basePath, step)
			dbg.Println("creating:", beforePath)
			if err := ioutil.WriteFile(beforePath, data, 0644); err != nil {
				warn.Printf("unable to create %q; %v", beforePath, err)
			}
		}
		after = func(g cfa.Graph, prim *primitive.Primitive) {
			data := []byte(dotAfterMerge(g, prim))
			afterPath := fmt.Sprintf("%s_xx_%04db.dot", basePath, step)
			dbg.Println("creating:", afterPath)
			if err := ioutil.WriteFile(afterPath, data, 0644); err != nil {
				warn.Printf("unable to create %q; %v", afterPath, err)
			}
			step++
		}
	}
	// Recovery control flow primitives.
	prims, err := hammock.Analyze(g, before, after)
	if err != nil {
		return errors.WithStack(err)
	}
	_ = prims
	// Print control flow graph.
	fmt.Println(g)
	return nil
}

// dotBeforeMerge returns the intermediate graph g in Graphviz DOT format with
// nodes before merge highlighted in red that are part of the located primitive.
func dotBeforeMerge(g cfa.Graph, prim *primitive.Primitive) string {
	// Colour nodes red.
	for _, dotID := range prim.Nodes {
		n, ok := g.NodeWithDOTID(dotID)
		if !ok {
			panic(fmt.Errorf("unable to locate node %q in control flow graph", dotID))
		}
		setFillColor(n, "red")
		if dotID == prim.Entry {
			// Add an external label with the name of the primitive for the entry
			// node of the primitive.
			// TODO: investigate whether quoting of attributes should be done by
			// gonum/encoding/dot.
			n.SetAttribute(encoding.Attribute{Key: "xlabel", Value: strconv.Quote(prim.Prim)})
		}
	}
	s := g.String()
	// Restore node colour.
	for _, dotID := range prim.Nodes {
		n, ok := g.NodeWithDOTID(dotID)
		if !ok {
			panic(fmt.Errorf("unable to locate node %q in control flow graph", dotID))
		}
		clearFillColor(n)
		if dotID == prim.Entry {
			// Clear external label from entry node of primitive.
			n.DelAttribute("xlabel")
		}
	}
	return s
}

// dotAfterMerge returns the intermediate graph g in Graphviz DOT format with
// the new node after merge highlighted in red that is part of the located
// primitive.
func dotAfterMerge(g cfa.Graph, prim *primitive.Primitive) string {
	// Colour nodes red.
	n, ok := g.NodeWithDOTID(prim.Entry)
	if !ok {
		panic(fmt.Errorf("unable to locate node %q in control flow graph", prim.Entry))
	}
	setFillColor(n, "red")
	s := g.String()
	// Restore node colour.
	clearFillColor(n)
	return s
}

// setFillColor sets the fillcolor attributes of the node to the given colour.
// The style attribute is also set to filled.
func setFillColor(n cfa.Node, color string) {
	n.SetAttribute(encoding.Attribute{Key: "fillcolor", Value: color})
	n.SetAttribute(encoding.Attribute{Key: "style", Value: "filled"})
}

// clearFillColor clears the fillcolor attribute of the node. The style
// attribute is also cleared.
func clearFillColor(n cfa.Node) {
	n.DelAttribute("fillcolor")
	n.DelAttribute("style")
}
