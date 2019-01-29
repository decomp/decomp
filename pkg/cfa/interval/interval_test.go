package interval

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/mewmew/lnp/pkg/cfa"
	"github.com/mewmew/lnp/pkg/cfg"
	"gonum.org/v1/gonum/graph"
)

func TestIntervals(t *testing.T) {
	golden := []struct {
		path string
		want [][]string
	}{
		{
			path: "testdata/sample.dot",
			want: [][]string{
				[]string{"B1", "B2", "B3", "B4", "B5"},
				[]string{"B6", "B7", "B8", "B9", "B10", "B11", "B12"},
				[]string{"B13", "B14", "B15"},
			},
		},
	}
	for _, gold := range golden {
		// Parse input.
		in := NewGraph()
		if err := cfg.ParseFileInto(gold.path, in); err != nil {
			t.Errorf("%q; unable to parse file; %v", gold.path, err)
			continue
		}
		// Locate intervals.
		is := Intervals(in)
		if len(is) != len(gold.want) {
			t.Errorf("%q: number of intervals mismatch; expected %d, got %d", gold.path, len(gold.want), len(is))
			continue
		}
		for i, want := range gold.want {
			var got []string
			// TODO: Update test to randomize node order. Then make sure the
			// intervals are calculated independent of what g.Nodes() returns. Use
			// reverse post-order.
			nodes := is[i].Nodes()
			for nodes.Next() {
				n := nodes.Node()
				nn, ok := n.(cfa.Node)
				if !ok {
					panic(fmt.Errorf("invalid node type; expected cfa.Node, got %T", n))
				}
				got = append(got, nn.DOTID())
			}
			sort.Strings(got)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%q; output mismatch; expected `%s`, got `%s`", gold.path, want, got)
				continue
			}
		}
	}
}

func TestCommonImmedDom(t *testing.T) {
	golden := []struct {
		path  string
		nodes []string
		want  string
	}{
		{
			path:  "testdata/cidom.dot",
			nodes: []string{"3", "4", "5"},
			want:  "1",
		},
	}
	for _, g := range golden {
		// Parse input.
		in := NewGraph()
		if err := cfg.ParseFileInto(g.path, in); err != nil {
			t.Errorf("%q; unable to parse file; %v", g.path, err)
			continue
		}
		// Get nodes with the specified node DOT IDs.
		nodes := NodesOf(in.Nodes())
		var ns []*Node
		for _, n := range nodes {
			if containsString(g.nodes, n.DOTID()) {
				ns = append(ns, n)
			}
		}
		// Locate common immediate dominator.
		dom := cfa.NewDom(in)
		cidom := commonImmedDom(ns, dom).(*Node)
		if g.want != cidom.DOTID() {
			t.Errorf("%q: common immediate dominator mismatch; expected %q, got %q", g.path, g.want, cidom.DOTID())
		}
	}
}

// containsString reports whether the slice contains the given string.
func containsString(ss []string, s string) bool {
	for _, t := range ss {
		if t == s {
			return true
		}
	}
	return false
}

// Assert that the interval implements the graph.Directed interface.
var _ graph.Directed = (*Interval)(nil)
