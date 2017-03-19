// TODO: Add support for parsing from standard input.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/decomp/decomp/cfg"
	"github.com/gonum/graph/encoding/dot"
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
	if err := restructure(dotPath, output, steps, indent); err != nil {
		log.Fatal(err)
	}
}

func restructure(dotPath, output string, steps, indent bool) error {
	g, err := cfg.ParseFile(dotPath)
	if err != nil {
		return errors.WithStack(err)
	}
	// TODO: Remove debug output.
	buf, err := dot.Marshal(g, "", "", "\t", false)
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Println(string(buf))
	return nil
}
