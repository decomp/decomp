package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	interval "github.com/mewmew/cifuentes_interval"
	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfg"
	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	for _, dotPath := range flag.Args() {
		if err := f(dotPath); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

func f(dotPath string) error {
	dst := interval.NewGraph()
	err := cfg.ParseFileInto(dotPath, dst)
	if err != nil {
		return errors.WithStack(err)
	}
	Gs, IIs := interval.DerivedSequence(dst)
	for i, g := range Gs {
		name := fmt.Sprintf("G_%d.dot", i+1)
		if err := ioutil.WriteFile(name, []byte(g.String()), 0644); err != nil {
			return errors.WithStack(err)
		}
	}
	for i, Is := range IIs {
		fmt.Printf("G_%d\n", i+1)
		for j, I := range Is {
			fmt.Printf("   I_%d\n", j+1)
			for nodes := I.Nodes(); nodes.Next(); {
				n := nodes.Node().(cfa.Node)
				fmt.Println("      n:", n.DOTID())
			}
		}
	}
	return nil
}
