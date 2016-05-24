// interval is a tool which locates intervals in graphs.
//
// Usage:
//    interval GRAPH...
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"decomp.org/x/graphs"
	"github.com/mewfork/dot"
	"github.com/mewkiz/pkg/errutil"
)

const use = `
Usage: interval GRAPH...
Locate each interval of two or more nodes in GRAPH (e.g. *.dot -> *.dot).
`

func usage() {
	fmt.Fprint(os.Stderr, use[1:])
}

func main() {
	flag.Usage = usage
	flag.Parse()
	for _, dotPath := range flag.Args() {
		is, err := locateIntervals(dotPath)
		if err != nil {
			log.Fatal(err)
		}
		for _, i := range is {
			iPath := i.Name + ".dot"
			buf := []byte(i.String())
			log.Printf("Creating %q", iPath)
			if err := ioutil.WriteFile(iPath, buf, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// locateIntervals locates all intervals in the provided graph.
func locateIntervals(dotPath string) ([]*graphs.Interval, error) {
	// Parse graph.
	g, err := dot.ParseFile(dotPath)
	if err != nil {
		return nil, errutil.Err(err)
	}

	// Locate intervals
	return graphs.GetIntervals(g), nil
}
