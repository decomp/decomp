// The ll2dot tool generates control flow graphs from LLVM IR (*.ll -> *.dot).
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/decomp/decomp/cfg"
	"github.com/gonum/graph/encoding/dot"
	"github.com/llir/llvm/asm"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
)

var dbg = log.New(os.Stderr, term.RedBold("ll2dot:")+" ", 0)

func usage() {
	const use = `
ll2dot [OPTION]... FILE...
Generate control flow graphs from LLVM IR (*.ll -> *.dot).
`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line flags.
	var (
		// fs represents a comma-separated list of function names.
		fs string
		// quiet specifies whether to suppress non-error messages.
		quiet bool
	)
	flag.StringVar(&fs, "funcs", "", "comma-separated list of function names")
	flag.BoolVar(&quiet, "q", false, "suppress non-error messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	funcs := make(map[string]bool)
	for _, f := range strings.Split(fs, ",") {
		if len(f) < 1 {
			continue
		}
		funcs[f] = true
	}
	if quiet {
		dbg.SetOutput(ioutil.Discard)
	}

	// Parse LLVM IR files.
	for _, llPath := range flag.Args() {
		if err := ll2dot(llPath, funcs); err != nil {
			log.Fatal(err)
		}
	}
}

func ll2dot(llPath string, funcs map[string]bool) error {
	dbg.Printf("parsing file %q.", llPath)
	module, err := asm.ParseFile(llPath)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range module.Funcs {
		if len(funcs) > 0 && !funcs[f.Name] {
			// Skip function if -funcs list exist and f is not part of it.
			dbg.Printf("skipping function %q not specified in -funcs flag.", f.Name)
			continue
		}
		dbg.Printf("generating CFG for function %q.", f.Name)
		g := cfg.New(f)
		buf, err := dot.Marshal(g, f.Name, "", "\t", false)
		if err != nil {
			return errors.WithStack(err)
		}
		fmt.Println(string(buf))
	}
	return nil
}
