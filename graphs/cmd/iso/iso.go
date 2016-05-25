//go:generate usagen iso
//go:generate mv z_usage.go z_usage.bak
//go:generate mango -plain iso.go
//go:generate mv z_usage.bak z_usage.go

// iso is a tool which locates subgraph isomorphisms in graphs.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"decomp.org/decomp/graphs"
	"decomp.org/decomp/graphs/iso"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/goutil"
	"github.com/mewkiz/pkg/osutil"
	"github.com/mewspring/dot"
)

// When flagStart is a non-empty string, locate an isomorphism of the subgraph
// in the graph which starts at the given node.
var flagStart string

func init() {
	flag.StringVar(&flagStart, "start", "", "Locate an isomorphism of SUB in GRAPH which starts at the given node.")
	flag.Usage = usage
}

const use = `
Usage: iso [OPTION]... SUB.dot GRAPH.dot
Locates isomorphisms of the subgraph SUB in GRAPH.

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
	err := locate(graphPath, subPath)
	if err != nil {
		log.Fatal(err)
	}
}

// locate parses the provided graphs and tries to locate isomorphisms of the
// subgraph in the graph.
func locate(graphPath, subPath string) error {
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

	// Locate isomorphisms.
	found := false
	if len(flagStart) > 0 {
		// Locate an isomorphism of sub in graph which starts at the node
		// specified by the "-start" flag.
		m, ok := iso.Isomorphism(graph, flagStart, sub)
		if ok {
			found = true
			printMapping(graph, sub, m)
		}
	} else {
		// Locate all isomorphisms of sub in graph.
		var names []string
		for name := range graph.Nodes.Lookup {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			m, ok := iso.Isomorphism(graph, name, sub)
			if !ok {
				continue
			}
			found = true
			printMapping(graph, sub, m)
		}
	}
	if !found {
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
