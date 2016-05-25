//go:generate usagen merge
//go:generate mv z_usage.go z_usage.bak
//go:generate mango -plain merge.go
//go:generate mv z_usage.bak z_usage.go

// merge is a tool which merges subgraph isomorphisms in graphs into single
// nodes.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"decomp.org/decomp/graphs"
	"decomp.org/decomp/graphs/iso"
	"decomp.org/decomp/graphs/merge"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/goutil"
	"github.com/mewkiz/pkg/osutil"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/mewspring/dot"
)

var (
	// When flagImage is true, generate an image representation of the CFG.
	flagImage bool
	// flagOut specifies the output path of the graph.
	flagOut string
	// When flagQuiet is true, suppress non-error messages.
	flagQuiet bool
	// When flagStart is a non-empty string, merge an isomorphism of the subgraph
	// in the graph which starts at the given node.
	flagStart string
)

func init() {
	flag.BoolVar(&flagImage, "img", false, "Generate an image representation of the CFG.")
	flag.StringVar(&flagOut, "o", "out.dot", "Output path of the graph.")
	flag.BoolVar(&flagQuiet, "q", false, "Suppress non-error messages.")
	flag.StringVar(&flagStart, "start", "", "Merge an isomorphism of SUB in GRAPH which starts at the given node.")
	flag.Usage = usage
}

const use = `
Usage: merge [OPTION]... SUB.dot GRAPH.dot
Merges isomorphisms of the subgraph SUB in GRAPH into single nodes.

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	subPath, graphPath := flag.Arg(0), flag.Arg(1)
	err := locateAndMerge(graphPath, subPath)
	if err != nil {
		log.Fatal(err)
	}
}

// locateAndMerge parses the provided graphs and tries to merge isomorphisms of
// the subgraph in the graph into single nodes.
func locateAndMerge(graphPath, subPath string) error {
	// Parse graph.
	graph, err := dot.ParseFile(graphPath)
	if err != nil {
		return errutil.Err(err)
	}

	// Search for subgraph in GOPATH if not found.
	if ok, _ := osutil.Exists(subPath); !ok {
		dir, err := goutil.SrcDir("decomp.org/decomp/graphs/testdata/primitives")
		if err != nil {
			return errutil.Err(err)
		}
		subPath = filepath.Join(dir, subPath)
	}
	sub, err := graphs.ParseSubGraph(subPath)
	if err != nil {
		return errutil.Err(err)
	}

	// Merge isomorphisms.
	found := false
	if len(flagStart) > 0 {
		// Merge an isomorphism of sub in graph which starts at the node
		// specified by the "-start" flag.
		m, ok := iso.Isomorphism(graph, flagStart, sub)
		if ok {
			found = true
			printMapping(graph, sub, m)
			_, err := merge.Merge(graph, m, sub)
			if err != nil {
				return errutil.Err(err)
			}
		}
	} else {
		// Merge all isomorphisms of sub in graph.
		for {
			m, ok := iso.Search(graph, sub)
			if !ok {
				break
			}
			found = true
			printMapping(graph, sub, m)
			_, err := merge.Merge(graph, m, sub)
			if err != nil {
				return errutil.Err(err)
			}
		}
	}

	// Store DOT and PNG representation of graph.
	if found {
		err = dump(graph)
		if err != nil {
			return errutil.Err(err)
		}
	} else {
		fmt.Println("not found.")
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
	fmt.Printf("Isomorphism of %q found at node %q:\n", sub.Name, entry)
	for _, sname := range snames {
		fmt.Printf("   %q=%q\n", sname, m[sname])
	}
}

// dump stores the graph as a DOT file and an image representation of the graph
// as a PNG file with filenames based on "-o" flag.
func dump(graph *dot.Graph) error {
	// Store graph to DOT file.
	dotPath := flagOut
	if !flagQuiet {
		log.Printf("Creating: %q\n", dotPath)
	}
	err := ioutil.WriteFile(dotPath, []byte(graph.String()), 0644)
	if err != nil {
		return errutil.Err(err)
	}

	// Generate an image representation of the graph.
	if flagImage {
		pngPath := pathutil.TrimExt(dotPath) + ".png"
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

	return nil
}
