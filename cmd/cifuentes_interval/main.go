package main

import (
	"flag"
	"fmt"
	"log"
	"sort"

	interval "github.com/mewmew/cifuentes_interval"
	"github.com/mewmew/lnp/pkg/cfg"
	"github.com/pkg/errors"
	"github.com/rickypai/natsort"
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
	/*
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
	*/
	interval.Analyze(dst, nil, nil)
	nodes := sortNodes(interval.NodesOf(dst.Nodes()))
	for _, n := range nodes {
		fmt.Println("node:      ", n.Node.DOTID())
		fmt.Println("preNum:    ", n.PreNum)
		fmt.Println("postNum:   ", n.PostNum)
		fmt.Println("revPostNum:", n.RevPostNum)
		fmt.Println("inLoop:    ", n.InLoop)
		fmt.Println("loopType:  ", n.LoopType)
		if n.LoopFollow != nil {
			fmt.Println("loopFollow:", n.LoopFollow.DOTID())
		}
		fmt.Println()
	}
	return nil
}

// sortNodes sorts the list of nodes by DOTID.
func sortNodes(ns []*interval.Node) []*interval.Node {
	less := func(i, j int) bool {
		return natsort.Less(ns[i].DOTID(), ns[j].DOTID())
	}
	sort.Slice(ns, less)
	return ns
}
