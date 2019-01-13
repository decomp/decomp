package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
	Gs, IIs := interval.DerivedSequence(dst)
	for i, g := range Gs {
		name := fmt.Sprintf("G_%d.dot", i+1)
		if err := ioutil.WriteFile(name, []byte(g.String()), 0644); err != nil {
			return errors.WithStack(err)
		}
	}
	for i, Is := range IIs {
		for j, I := range Is {
			name := fmt.Sprintf("I_%d_%d.dot", i+1, j+1)
			if err := ioutil.WriteFile(name, []byte(I.String()), 0644); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	prims := interval.Analyze(dst, nil, nil)
	fmt.Println("=== [ graph nodes ] ========")
	printNodes(interval.NodesOf(dst.Nodes()))
	buf, err := json.MarshalIndent(prims, "", "\t")
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Println(string(buf))
	return nil
}

func printNodes(ns []*interval.Node) {
	sortNodes(ns)
	for _, n := range ns {
		fmt.Printf("Node:       %v   (pre: %d, rev-post: %d)\n", n.Node.DOTID(), n.PreNum, n.RevPostNum)
		if n.LoopHead != nil {
			fmt.Println("LoopHead:  ", n.LoopHead.DOTID())
		}
		if n.LoopType != interval.LoopTypeNone {
			fmt.Println("LoopType:  ", n.LoopType)
		}
		if n.LoopFollow != nil {
			fmt.Println("LoopFollow:", n.LoopFollow.DOTID())
		}
		if n.Follow != nil {
			fmt.Println("Follow:    ", n.Follow.DOTID())
		}
		if n.IsCondNode != false {
			fmt.Println("IsCondNode:", n.IsCondNode)
		}
		if len(n.CompCond) > 0 {
			fmt.Println("CompCond:  ", n.CompCond)
		}
		fmt.Println()
	}
}

// sortNodes sorts the list of nodes by DOTID.
func sortNodes(ns []*interval.Node) []*interval.Node {
	less := func(i, j int) bool {
		return natsort.Less(ns[i].DOTID(), ns[j].DOTID())
	}
	sort.Slice(ns, less)
	return ns
}
