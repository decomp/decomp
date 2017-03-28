// TODO: Add support for parsing from standard input.

// The restructure tool recovers control flow primitives from DOT control flow
// graphs (*.dot -> *.json).
//
// The input of restructure is a Graphviz DOT file, containing the unstructured
// control flow graph of a function, and the output is a JSON stream describing
// how the recovered high-level control flow primitives relate to the nodes of
// the control flow graph.
//
// Usage:
//
//    restructure [OPTION]... FILE.dot
//
// Flags:
//
//    -entry string
//          entry node of the control flow graph
//    -indent
//          indent JSON output
//    -o string
//          output path
//    -q    suppress non-error messages
//    -steps
//      	output intermediate control flow graphs at each step
package main

import (
	"bytes"
	"encoding/json"
	goerrors "errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/decomp/decomp/cfa"
	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
	"github.com/gonum/graph"
	"github.com/gonum/graph/encoding/dot"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
)

// dbg represents a logger with the "restructure:" prefix, which logs debug
// messages to standard error.
var dbg = log.New(os.Stderr, term.MagentaBold("restructure:")+" ", 0)

func usage() {
	const use = `
Recover control flow primitives from DOT control flow graphs (*.dot -> *.json).

Usage:

	restructure [OPTION]... FILE.dot

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
	name := pathutil.FileName(dotPath)
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
	prims, err := restructure(g, entry, steps, name)
	if err != nil {
		if errors.Cause(err) == ErrIncomplete {
			// Do _not_ terminate on incomplete control flow recovery. Instead
			// print partial results.
			dbg.Printf("WARNING: %v", err)
		} else {
			log.Fatalf("%+v", err)
		}
	}

	// Store JSON output.
	w := os.Stdout
	if len(output) > 0 {
		f, err := os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w = f
	}
	if err := writeJSON(w, prims, indent); err != nil {
		log.Fatalf("%+v", err)
	}
}

// locateEntryNode attempts to locate the entry node of the control flow graph,
// either by label if specified, or by searching for a single node in the
// control flow graph with no incoming edges.
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

// ErrIncomplete signals an incomplete control flow recovery.
var ErrIncomplete = goerrors.New("incomplete control flow recovery")

// restructure attempts to recover the control flow primitives of a given
// control flow graph. It does so by repeatedly locating and merging structured
// subgraphs (graph representations of control flow primitives) into single
// nodes until the entire graph is reduced into a single node or no structured
// subgraphs may be located. The steps argument specifies whether to record the
// intermediate CFGs at each step. The returned list of primitives is ordered in
// the same sequence as they were located.
func restructure(g *cfg.Graph, entry graph.Node, steps bool, name string) ([]*primitive.Primitive, error) {
	prims := make([]*primitive.Primitive, 0)
	// Locate control flow primitives.
	for step := 1; len(g.Nodes()) > 1; step++ {
		// Locate primitive.
		dom := cfg.NewDom(g, entry)
		prim, err := cfa.FindPrim(g, dom)
		if err != nil {
			return prims, errors.Wrap(ErrIncomplete, err.Error())
		}
		prims = append(prims, prim)

		// Output pre-merge intermediate CFG.
		if steps {
			path := fmt.Sprintf("%s_%04da.dot", name, step)
			var highlight []string
			for _, node := range prim.Nodes {
				highlight = append(highlight, node)
			}
			if err := storeStep(g, name, path, highlight); err != nil {
				return nil, errors.WithStack(err)
			}
		}

		// Merge the nodes of the primitive into a single node.
		if err := cfa.Merge(g, prim); err != nil {
			return nil, errors.WithStack(err)
		}
		// Handle special case where entry node has been replaced by primitive
		// node.
		if !g.Has(entry) {
			entry = g.NodeByLabel(prim.Entry)
			if entry == nil {
				return nil, errors.Errorf("unable to locate entry node %q", prim.Entry)
			}
		}

		// Output post-merge intermediate CFG.
		if steps {
			path := fmt.Sprintf("%s_%04db.dot", name, step)
			highlight := []string{prim.Entry}
			if err := storeStep(g, name, path, highlight); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}
	return prims, nil
}

// storeStep stores a DOT representation of g to path with the specified nodes
// highlighted in red.
func storeStep(g *cfg.Graph, name, path string, highlight []string) error {
	for _, h := range highlight {
		n := g.NodeByLabel(h)
		if n == nil {
			return errors.Errorf("unable to located node %q to be highlighted", h)
		}
		n.Attrs["style"] = "filled"
		n.Attrs["fillcolor"] = "red"
	}
	buf, err := dot.Marshal(g, name, "", "\t", false)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, h := range highlight {
		n := g.NodeByLabel(h)
		if n == nil {
			return errors.Errorf("unable to located node %q to be highlighted", h)
		}
		delete(n.Attrs, "style")
		delete(n.Attrs, "fillcolor")
	}
	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// writeJSON writes the primitives in JSON format to w.
func writeJSON(w io.Writer, prims []*primitive.Primitive, indent bool) error {
	// Output indented JSON.
	if indent {
		buf, err := json.MarshalIndent(prims, "", "\t")
		if err != nil {
			return errors.WithStack(err)
		}
		buf = append(buf, '\n')
		if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	// Output JSON.
	enc := json.NewEncoder(w)
	if err := enc.Encode(prims); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// label returns the label of the node.
func label(n graph.Node) string {
	if n, ok := n.(*cfg.Node); ok {
		return n.Label
	}
	panic(fmt.Sprintf("invalid node type; expected *cfg.Node, got %T", n))
}
