package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mewmew/lnp/cfa"
	"github.com/mewmew/lnp/cfa/hammock"
	"github.com/mewmew/lnp/cfa/primitive"
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
	before := func(g cfa.Graph, prim *primitive.Primitive) {
		fmt.Println("before merge:", g)
		fmt.Println("located prim:", prim)
	}
	after := func(g cfa.Graph, prim *primitive.Primitive) {
		fmt.Println("after merge:", g)
	}
	prims, err := hammock.Analyze(g, before, after)
	if err != nil {
		return errors.WithStack(err)
	}
	_ = prims
	// Print control flow graph.
	fmt.Println(g)
	return nil
}
