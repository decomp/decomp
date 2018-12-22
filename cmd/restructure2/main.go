package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mewmew/lnp/cfa/hammock"
	"github.com/mewmew/lnp/cfg"
	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	dotPath := flag.Arg(0)
	if err := restructure(dotPath); err != nil {
		log.Fatalf("%+v", err)
	}
}

func restructure(dotPath string) error {
	// Parse control flow graph.
	g, err := cfg.ParseFile(dotPath)
	if err != nil {
		return errors.WithStack(err)
	}
	// Recovery control flow primitives.
	prims, err := hammock.Analyze(g)
	if err != nil {
		return errors.WithStack(err)
	}
	_ = prims
	// Print control flow graph.
	fmt.Println(g)
	return nil
}
