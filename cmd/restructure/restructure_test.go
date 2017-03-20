//+build ignore

// TODO: Figure out how to handle non-determinism of node mapping in if-else
// primitive.
//
// Currently, the edge information (true/false-branch) is not taken into account
// (and probably should not be, as a loop may break on either true or false).
// However, as a result, the following are undifferentiatable.
//
//    if (A) {
//       B
//    } else {
//       C
//    }
//    D
//
//    if (A) {
//       C
//    } else {
//       B
//    }
//    D

package main

import (
	"reflect"
	"testing"

	"github.com/decomp/decomp/cfa/primitive"
	"github.com/decomp/decomp/graph/cfg"
)

func TestRestructure(t *testing.T) {
	golden := []struct {
		path  string
		entry string
		want  []*primitive.Primitive
	}{
		{
			path:  "testdata/if-else.dot",
			entry: "2",
			want: []*primitive.Primitive{
				{
					Prim: "if_else",
					Node: "if_else_0",
					Nodes: map[string]string{
						"cond":       "2",
						"body_true":  "4",
						"body_false": "5",
						"exit":       "6",
					},
					Entry: "2",
					Exit:  "6",
				},
			},
		},
		/*{
			path:  "testdata/stmt.dot",
			entry: "0",
			want: []*primitive.Primitive{
				{
					Prim: "pre_loop",
					Node: "pre_loop_0",
					Nodes: map[string]string{
						"cond": "89",
						"body": "92",
						"exit": "93",
					},
					Entry: "89",
					Exit:  "93",
				},
				{
					Prim: "if",
					Node: "if_0",
					Nodes: map[string]string{
						"cond": "71",
						"body": "74",
						"exit": "75",
					},
					Entry: "71",
					Exit:  "75",
				},
				{
					Prim: "if",
					Node: "if_1",
					Nodes: map[string]string{
						"cond": "17",
						"body": "24",
						"exit": "32",
					},
					Entry: "17",
					Exit:  "32",
				},
				{
					Prim: "if_return",
					Node: "if_return_0",
					Nodes: map[string]string{
						"cond": "48",
						"body": "52",
						"exit": "51",
					},
					Entry: "48",
					Exit:  "51",
				},
				{
					Prim: "if_return",
					Node: "if_return_1",
					Nodes: map[string]string{
						"cond": "98",
						"body": "102",
						"exit": "101",
					},
					Entry: "98",
					Exit:  "101",
				},
				{
					Prim: "if_return",
					Node: "if_return_2",
					Nodes: map[string]string{
						"cond": "3",
						"body": "7",
						"exit": "6",
					},
					Entry: "3",
					Exit:  "6",
				},
				{
					Prim: "if_return",
					Node: "if_return_3",
					Nodes: map[string]string{
						"cond": "10",
						"body": "14",
						"exit": "13",
					},
					Entry: "10",
					Exit:  "13",
				},
				{
					Prim: "if_return",
					Node: "if_return_4",
					Nodes: map[string]string{
						"cond": "if_0",
						"body": "81",
						"exit": "80",
					},
					Entry: "if_0",
					Exit:  "80",
				},
				{
					Prim: "if_return",
					Node: "if_return_5",
					Nodes: map[string]string{
						"cond": "39",
						"body": "45",
						"exit": "44",
					},
					Entry: "39",
					Exit:  "44",
				},
				{
					Prim: "seq",
					Node: "seq_0",
					Nodes: map[string]string{
						"entry": "if_return_4",
						"exit":  "84",
					},
					Entry: "if_return_4",
					Exit:  "84",
				},
				{
					Prim: "seq",
					Node: "seq_1",
					Nodes: map[string]string{
						"entry": "88",
						"exit":  "pre_loop_0",
					},
					Entry: "88",
					Exit:  "pre_loop_0",
				},
				{
					Prim: "seq",
					Node: "seq_2",
					Nodes: map[string]string{
						"entry": "if_return_3",
						"exit":  "if_1",
					},
					Entry: "if_return_3",
					Exit:  "if_1",
				},
				{
					Prim: "seq",
					Node: "seq_3",
					Nodes: map[string]string{
						"entry": "if_return_5",
						"exit":  "if_return_0",
					},
					Entry: "if_return_5",
					Exit:  "if_return_0",
				},
				{
					Prim: "seq",
					Node: "seq_4",
					Nodes: map[string]string{
						"entry": "seq_3",
						"exit":  "55",
					},
					Entry: "seq_3",
					Exit:  "55",
				},
				{
					Prim: "seq",
					Node: "seq_3",
					Nodes: map[string]string{
						"entry": "if_return_2",
						"exit":  "seq_2",
					},
					Entry: "if_return_2",
					Exit:  "seq_2",
				},
				{
					Prim: "seq",
					Node: "seq_2",
					Nodes: map[string]string{
						"entry": "if_return_1",
						"exit":  "105",
					},
					Entry: "if_return_1",
					Exit:  "105",
				},
				{
					Prim: "if_else",
					Node: "if_else_0",
					Nodes: map[string]string{
						"cond":       "94",
						"body_true":  "97",
						"body_false": "seq_2",
						"exit":       "106",
					},
					Entry: "94",
					Exit:  "106",
				},
				{
					Prim: "if_else",
					Node: "if_else_1",
					Nodes: map[string]string{
						"cond":       "85",
						"body_true":  "seq_1",
						"body_false": "if_else_0",
						"exit":       "107",
					},
					Entry: "85",
					Exit:  "107",
				},
				{
					Prim: "if_else",
					Node: "if_else_0",
					Nodes: map[string]string{
						"cond":       "68",
						"body_true":  "if_else_1",
						"body_false": "seq_0",
						"exit":       "108",
					},
					Entry: "68",
					Exit:  "108",
				},
				{
					Prim: "if_else",
					Node: "if_else_1",
					Nodes: map[string]string{
						"cond":       "36",
						"body_true":  "if_else_0",
						"body_false": "seq_4",
						"exit":       "109",
					},
					Entry: "36",
					Exit:  "109",
				},
				{
					Prim: "if_else",
					Node: "if_else_0",
					Nodes: map[string]string{
						"cond":       "0",
						"body_true":  "seq_3",
						"body_false": "if_else_1",
						"exit":       "110",
					},
					Entry: "0",
					Exit:  "110",
				},
			},
		},*/
	}

	for _, gold := range golden {
		g, err := cfg.ParseFile(gold.path)
		if err != nil {
			t.Errorf("%q: unable to parse DOT file; %v", gold.path, err)
			continue
		}

		// Locate entry node.
		entry, err := locateEntryNode(g, gold.entry)
		if err != nil {
			t.Errorf("%q: unable to locate entry node; %v", gold.path, err)
		}

		got, err := restructure(g, entry, false, "")
		if err != nil {
			t.Errorf("%q: unable to restructure; %v", gold.path, err)
			continue
		}
		if !reflect.DeepEqual(got, gold.want) {
			t.Errorf("%q: primitive mismatch; expected %#v, got %#v", gold.path, gold.want[0], got[0])
		}
	}
}
