// restructure is a tool which recovers high-level control flow primitives from
// control flow graphs (e.g. *.dot -> *.json). It takes an unstructured CFG (in
// Graphviz DOT file format) as input and produces a structured CFG (in JSON),
// which describes how the high-level control flow primitives relate to the
// nodes of the CFG.
//
// Usage:
//     restructure [OPTION]... [CFG.dot]
//
//     Flags:
//       -img
//             Generate image representations of the intermediate CFGs.
//       -indent
//             Indent JSON output.
//       -o string
//             Output path.
//       -prims string
//             An ordered, comma-separated list of control flow primitives (*.dot). Restructure
//             searches for missing files in $GOPATH/src/decomp.org/x/cmd/restructure/primitives/.
//             (default "pre_loop.dot,post_loop.dot,list.dot,if.dot,if_else.dot,if_return.dot")
//       -q    Suppress non-error messages.
//       -steps
//             Output intermediate CFGs at each step.
//       -v    Verbose output.
//
// Example input:
//    digraph foo {
//       E -> F
//       E -> H
//       F -> G
//       G -> H
//       E [label="entry"]
//       F
//       G
//       H [label="exit"]
//    }
//
// Example output:
//    [
//       {
//          "prim": "list",
//          "node": "list0",
//          "nodes": {
//             "entry": "F",
//             "exit": "G"
//          }
//       },
//       {
//          "prim": "if",
//          "node": "if0",
//          "nodes": {
//             "cond": "E",
//             "body": "list0",
//             "exit": "H"
//          }
//       },
//    ]
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"decomp.org/x/graphs"
	"decomp.org/x/graphs/iso"
	"decomp.org/x/graphs/merge"
	"decomp.org/x/graphs/primitive"
	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/goutil"
	"github.com/mewkiz/pkg/osutil"
)

var (
	// When flagImage is true, generate image representations of the intermediate
	// CFGs.
	flagImage bool
	// When flagIndent is true, indent JSON output.
	flagIndent bool
	// flagOutput specifies the output path.
	flagOutput string
	// flagPrimitives is an ordered, comma-separated list of control flow
	// primitives (*.dot). Restructure searches for missing files in
	// $GOPATH/src/decomp.org/x/cmd/restructure/primitives/.
	flagPrimitives string
	// When flagQuiet is true, suppress non-error messages.
	flagQuiet bool
	// When flagSteps is true, output intermediate control flow graphs at each
	// step.
	flagSteps bool
	// When flagVerbose is true, enable verbose output.
	flagVerbose bool
)

func init() {
	flag.BoolVar(&flagImage, "img", false, "Generate image representations of the intermediate CFGs.")
	flag.BoolVar(&flagIndent, "indent", false, "Indent JSON output.")
	flag.StringVar(&flagOutput, "o", "", "Output path.")
	flag.StringVar(&flagPrimitives, "prims", "pre_loop.dot,post_loop.dot,list.dot,if.dot,if_else.dot,if_return.dot", "An ordered, comma-separated list of control flow primitives (*.dot). Restructure searches for missing files in $GOPATH/src/decomp.org/x/cmd/restructure/primitives/.")
	flag.BoolVar(&flagQuiet, "q", false, "Suppress non-error messages.")
	flag.BoolVar(&flagSteps, "steps", false, "Output intermediate CFGs at each step.")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose output.")
	flag.Usage = usage
}

const use = `
Usage: restructure [OPTION]... [CFG.dot]
Recover control flow primitives from control flow graphs (e.g. *.dot -> *.json).
`

func usage() {
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if flagImage && !flagSteps {
		log.Fatalf(`invalid use of "-img" flag; must be used in conjuction with "-steps"`)
	}

	var dotPath string
	switch flag.NArg() {
	case 0:
		// Read from stdin.
		dotPath = "-"
	case 1:
		// Read from FILE.
		dotPath = flag.Arg(0)
	default:
		flag.Usage()
		os.Exit(1)
	}

	// Parse the graph representations of the high-level control flow primitives.
	//
	// subs is an ordered list of subgraphs representing common control-flow
	// primitives such as 2-way conditionals, pre-test loops, etc.
	subs, err := parseSubs(strings.Split(flagPrimitives, ","))
	if err != nil {
		log.Fatal(err)
	}

	// Create a structured CFG from the unstructured CFG.
	prims, err := restructure(dotPath, subs)
	if err != nil {
		log.Fatal(err)
	}

	// Print the JSON output to stdout or the path specified by the "-o" flag.
	w := os.Stdout
	if len(flagOutput) > 0 {
		f, err := os.Create(flagOutput)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w = f
	}
	if err := writeJSON(w, prims); err != nil {
		log.Fatal(err)
	}
}

// parseSubs parses the graph representations of the given high-level control
// flow primitives. If unable to locate a subgraph, a second attempt is made by
// prepending $GOPATH/src/decomp.org/x/cmd/restructure/primitives/ to the
// subgraph path.
func parseSubs(subPaths []string) (subs []*graphs.SubGraph, err error) {
	// Prepend $GOPATH/src/decomp.org/x/cmd/restructure/primitives/ to the path
	// of missing subgraphs.
	subDir, err := goutil.SrcDir("decomp.org/x/cmd/restructure/primitives")
	if err != nil {
		return nil, errutil.Err(err)
	}
	for i, subPath := range subPaths {
		if ok, _ := osutil.Exists(subPath); !ok {
			subPath = filepath.Join(subDir, subPath)
			subPaths[i] = subPath
		}
	}

	// Parse subgraphs representing control flow primitives.
	for _, subPath := range subPaths {
		sub, err := graphs.ParseSubGraph(subPath)
		if err != nil {
			return nil, errutil.Err(err)
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

// restructure attempts to recover the control flow primitives of a given
// control flow graph. It does so by repeatedly locating and merging structured
// subgraphs (graph representations of control flow primitives) into single
// nodes until the entire graph is reduced into a single node or no structured
// subgraphs may be located. The list of primitives is ordered in the same
// sequence as they were located.
func restructure(dotPath string, subs []*graphs.SubGraph) (prims []*primitive.Primitive, err error) {
	// Parse the unstructured CFG.
	var graph *dot.Graph
	switch dotPath {
	case "-":
		// Read from stdin.
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, errutil.Err(err)
		}
		graph, err = dot.Read(buf)
		if err != nil {
			return nil, errutil.Err(err)
		}
	default:
		// Read from FILE.
		graph, err = dot.ParseFile(dotPath)
		if err != nil {
			return nil, errutil.Err(err)
		}
	}
	if len(graph.Nodes.Nodes) == 0 {
		return nil, errutil.Newf("unable to restructure empty graph %q", dotPath)
	}

	// Locate control flow primitives.
	for step := 1; len(graph.Nodes.Nodes) > 1; step++ {
		prim, err := findPrim(graph, subs, step)
		if err != nil {
			return nil, errutil.Err(err)
		}
		prims = append(prims, prim)
	}

	return prims, nil
}

// findPrim locates a control flow primitive in the provided control flow graph
// and merges its nodes into a single node.
func findPrim(graph *dot.Graph, subs []*graphs.SubGraph, step int) (*primitive.Primitive, error) {
	for _, sub := range subs {
		// Locate an isomorphism of sub in graph.
		m, ok := iso.Search(graph, sub)
		if !ok {
			// No match, try next control flow primitive.
			continue
		}
		if flagVerbose && !flagQuiet {
			printMapping(graph, sub, m)
		}

		// Output pre-merge intermediate control flow graphs.
		if flagSteps {
			// Highlight nodes to be replaced in red.
			for _, preNodeName := range m {
				preNode, ok := graph.Nodes.Lookup[preNodeName]
				if !ok {
					return nil, errutil.Newf("unable to locate pre-merge node %q", preNodeName)
				}
				if preNode.Attrs == nil {
					preNode.Attrs = dot.NewAttrs()
				}
				preNode.Attrs["fillcolor"] = "red"
				preNode.Attrs["style"] = "filled"
			}

			// Store pre-merge DOT graph.
			stepName := fmt.Sprintf("%s_%da", graph.Name, step)
			if err := createDOT(stepName, graph); err != nil {
				return nil, errutil.Err(err)
			}

			// Restore node colour.
			for _, preNodeName := range m {
				preNode, ok := graph.Nodes.Lookup[preNodeName]
				if !ok {
					return nil, errutil.Newf("unable to locate pre-merge node %q", preNodeName)
				}

				delete(preNode.Attrs, "fillcolor")
				delete(preNode.Attrs, "style")
			}
		}

		// Check if one of the nodes to be merged has the label "entry".
		hasEntry := false
		for _, preNodeName := range m {
			preNode, ok := graph.Nodes.Lookup[preNodeName]
			if !ok {
				return nil, errutil.Newf("unable to locate pre-merge node %q", preNodeName)
			}
			if preNode.Attrs != nil && preNode.Attrs["label"] == "entry" {
				hasEntry = true
				break
			}
		}

		// Merge the nodes of the subgraph isomorphism into a single node.
		postNodeName, err := merge.Merge(graph, m, sub)
		if err != nil {
			return nil, errutil.Err(err)
		}

		// Add "entry" label to new node if present in the pre-merge nodes.
		if hasEntry {
			postNode, ok := graph.Nodes.Lookup[postNodeName]
			if !ok {
				return nil, errutil.Newf("unable to locate post-merge node %q", postNodeName)
			}
			if postNode.Attrs == nil {
				postNode.Attrs = dot.NewAttrs()
			}
			postNode.Attrs["label"] = "entry"
			index := postNode.Index
			graph.Nodes.Nodes[0], graph.Nodes.Nodes[index] = postNode, graph.Nodes.Nodes[0]
		}

		// Output post-merge intermediate control flow graphs.
		if flagSteps {
			// Highlight node to be replaced in red.
			postNode, ok := graph.Nodes.Lookup[postNodeName]
			if !ok {
				return nil, errutil.Newf("unable to locate post-merge node %q", postNodeName)
			}
			if postNode.Attrs == nil {
				postNode.Attrs = dot.NewAttrs()
			}
			postNode.Attrs["fillcolor"] = "red"
			postNode.Attrs["style"] = "filled"

			// Store post-merge DOT graph.
			stepName := fmt.Sprintf("%s_%db", graph.Name, step)
			if err := createDOT(stepName, graph); err != nil {
				return nil, errutil.Err(err)
			}

			// Restore node colour.
			delete(postNode.Attrs, "fillcolor")
			delete(postNode.Attrs, "style")
		}

		// Create a new control flow primitive.
		prim := &primitive.Primitive{
			Node:  postNodeName,
			Prim:  sub.Name,
			Nodes: m,
		}
		return prim, nil
	}

	return nil, errutil.New("unable to locate control flow primitive")
}

// createDOT creates a DOT graph with the given file name.
func createDOT(stepName string, graph *dot.Graph) error {
	dotPath := stepName + ".dot"
	if !flagQuiet {
		log.Printf("Creating: %q", dotPath)
	}
	buf := []byte(graph.String())
	if err := ioutil.WriteFile(dotPath, buf, 0644); err != nil {
		return errutil.Err(err)
	}

	// Generate an image representation of the control flow graph.
	if flagImage {
		pngPath := stepName + ".png"
		if !flagQuiet {
			log.Printf("Creating: %q", pngPath)
		}
		cmd := exec.Command("dot", "-Tpng", "-o", pngPath, dotPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errutil.Err(err)
		}
	}
	return nil
}

// writeJSON writes the primitives in JSON format to w.
func writeJSON(w io.Writer, prims []*primitive.Primitive) error {
	if flagIndent {
		buf, err := json.MarshalIndent(prims, "", "\t")
		if err != nil {
			return errutil.Err(err)
		}
		_, err = io.Copy(w, bytes.NewReader(buf))
		if err != nil {
			return errutil.Err(err)
		}
		return nil
	}
	enc := json.NewEncoder(w)
	err := enc.Encode(prims)
	if err != nil {
		return errutil.Err(err)
	}
	return nil
}

// printMapping prints the mapping from sub node name to graph node name for an
// isomorphism of sub in graph.
func printMapping(graph *dot.Graph, sub *graphs.SubGraph, m map[string]string) {
	entry := m[sub.Entry()]
	var snames []string
	for sname := range m {
		snames = append(snames, sname)
	}
	sort.Strings(snames)
	fmt.Fprintf(os.Stderr, "Isomorphism of %q found at node %q:\n", sub.Name, entry)
	for _, sname := range snames {
		fmt.Fprintf(os.Stderr, "   %q=%q\n", sname, m[sname])
	}
}
