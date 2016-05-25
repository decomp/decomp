package main

import (
	"reflect"
	"testing"

	"decomp.org/decomp/graphs/primitive"
)

func TestRestructure(t *testing.T) {
	golden := []struct {
		path string
		want []*primitive.Primitive
	}{
		{
			path: "testdata/foo.dot",
			want: []*primitive.Primitive{
				{
					Prim:  "list",
					Node:  "list0",
					Nodes: map[string]string{"entry": "F", "exit": "G"},
				},
				{
					Prim:  "if",
					Node:  "if0",
					Nodes: map[string]string{"cond": "E", "body": "list0", "exit": "H"},
				},
			},
		},
		{
			path: "testdata/bar.dot",
			want: []*primitive.Primitive{
				{
					Prim:  "if_else",
					Node:  "if_else0",
					Nodes: map[string]string{"cond": "F", "body_true": "H", "body_false": "G", "exit": "I"},
				},
				{
					Prim:  "pre_loop",
					Node:  "pre_loop0",
					Nodes: map[string]string{"cond": "E", "body": "if_else0", "exit": "J"},
				},
			},
		},
	}

	// Parse the graph representations of the high-level control flow primitives.
	//
	// subs is an ordered list of subgraphs representing common control-flow
	// primitives such as 2-way conditionals, pre-test loops, etc.
	subs, err := parseSubs([]string{"pre_loop.dot", "post_loop.dot", "list.dot", "if.dot", "if_else.dot", "if_return.dot"})
	if err != nil {
		t.Fatal(err)
	}

	for i, g := range golden {
		got, err := restructure(g.path, subs)
		if err != nil {
			t.Errorf("i=%d: error; %v", i, err)
			continue
		}
		if !reflect.DeepEqual(got, g.want) {
			t.Errorf("i=%d: primitive mismatch; expected %#v, got %#v", i, g.want[0], got[0])
		}
	}
}
