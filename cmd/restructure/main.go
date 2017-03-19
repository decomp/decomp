// TODO: Add support for parsing from standard input.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/decomp/decomp/graph/cfg"
	"github.com/gonum/graph"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
)

// dbg represents a logger with the "restructure:" prefix, which logs debug
// messages to standard error.
var dbg = log.New(os.Stderr, term.MagentaBold("restructure:")+" ", 0)

func usage() {
	const use = `
restructure [OPTION]... FILE.dot
Recover control flow primitives from DOT control flow graphs (*.dot -> *.json).

Flags:
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line flags.
	var (
		// entryLabel specifies the entry node of the control flow graph.
		entryLabel string
		// indent specifies whether to indent JSON output.
		indent bool
		// output specifies the output path.
		output string
		// quiet specifies whether to suppress non-error messages.
		quiet bool
		// steps specifies whether to output intermediate control flow graphs at
		// each step.
		steps bool
	)
	flag.StringVar(&entryLabel, "entry", "", "entry node of the control flow graph")
	flag.BoolVar(&indent, "indent", false, "indent JSON output")
	flag.StringVar(&output, "o", "", "output path")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.BoolVar(&steps, "steps", false, "output intermediate control flow graphs at each step")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	// Mute debug messages if `-q` is set.
	if quiet {
		dbg.SetOutput(ioutil.Discard)
	}

	// Parse DOT file.
	dotPath := flag.Arg(0)
	g, err := cfg.ParseFile(dotPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// Locate entry node.
	entry, err := locateEntryNode(g, entryLabel)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// Perform control flow analysis.
	if err := restructure(g, entry, steps); err != nil {
		log.Fatalf("%+v", err)
	}
}

func locateEntryNode(g *cfg.Graph, entryLabel string) (graph.Node, error) {
	if len(entryLabel) > 0 {
		entry := g.NodeByLabel(entryLabel)
		if entry == nil {
			return nil, errors.Errorf("unable to locate entry node with node label %q", entryLabel)
		}
		return entry, nil
	}
	var entry graph.Node
	for _, n := range g.Nodes() {
		preds := g.To(n)
		if len(preds) == 0 {
			if entry != nil {
				return nil, errors.Errorf("more than one candidate for the entry node located; prev %#v, new %#v", entry, n)
			}
			entry = n
		}
	}
	if entry == nil {
		return nil, errors.Errorf("unable to locate entry node; try specifying an entry node label using the -entry flag")
	}
	return entry, nil
}

func restructure(g *cfg.Graph, entry graph.Node, steps bool) error {
	return nil
}
