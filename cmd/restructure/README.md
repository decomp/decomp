## WIP

This project is a *work in progress*. The implementation is *incomplete* and subject to change. The documentation can be inaccurate.

# restructure

[![Build Status](https://travis-ci.org/decomp/restructure.svg?branch=master)](https://travis-ci.org/decomp/restructure)
[![Coverage Status](https://img.shields.io/coveralls/decomp/restructure.svg)](https://coveralls.io/r/decomp/restructure?branch=master)
[![GoDoc](https://godoc.org/decomp.org/x/cmd/restructure?status.svg)](https://godoc.org/decomp.org/x/cmd/restructure)

`restructure` is a tool which recovers high-level control flow primitives from control flow graphs (e.g. *.dot -> *.json). It takes an unstructured CFG (in Graphviz DOT file format) as input and produces a structured CFG (in JSON), which describes how the high-level control flow primitives relate to the nodes of the CFG.

## Installation

```shell
go get decomp.org/x/cmd/restructure
```

## Usage

```
restructure [OPTION]... [CFG.dot]

Flags:
  -img
        Generate image representations of the intermediate CFGs.
  -indent
        Indent JSON output.
  -o string
        Output path.
  -prims string
        An ordered, comma-separated list of control flow primitives (*.dot). Restructure
        searches for missing files in $GOPATH/src/decomp.org/x/cmd/restructure/primitives/.
        (default "pre_loop.dot,post_loop.dot,list.dot,if.dot,if_else.dot,if_return.dot")
  -q    Suppress non-error messages.
  -steps
        Output intermediate CFGs at each step.
  -v    Verbose output.
```

## Examples

### Control flow recovery

Recover the high-level control flow primitives from the control flow graph [foo.dot](testdata/foo.dot).

```bash
$ restructure -v -indent foo.dot
// Stderr output:
// Isomorphism of "list" found at node "F":
//    "entry"="F"
//    "exit"="G"
// Isomorphism of "if" found at node "E":
//    "body"="list0"
//    "cond"="E"
//    "exit"="H"
//
// Output:
[
    {
        "prim": "list",
        "node": "list0",
        "nodes": {
            "entry": "F",
            "exit": "G"
        }
    },
    {
        "prim": "if",
        "node": "if0",
        "nodes": {
            "body": "list0",
            "cond": "E",
            "exit": "H"
        }
    }
]
```

INPUT:
* [foo.dot](testdata/foo.dot): unstructured control flow graph.

![foo.dot subgraph](https://raw.githubusercontent.com/decomp/restructure/master/testdata/foo.png)

OUTPUT:
* [foo.json](testdata/foo.json): structured control flow graph.

```c
if E {
    F
    G
}
H
```

### Output intermediate CFGs

Output intermediate CFGs at each step of the control flow recovery, when analysing the CFG [stmt.dot].

[stmt.dot]: https://raw.githubusercontent.com/decomp/graphs/master/testdata/c4_graphs/stmt.dot

```bash
$ restructure -steps -img stmt.dot
// Output to stderr:
// 2015/05/26 15:50:43 Creating: "stmt_1a.dot"
// 2015/05/26 15:50:43 Creating: "stmt_1a.png"
// 2015/05/26 15:50:43 Creating: "stmt_1b.dot"
// 2015/05/26 15:50:43 Creating: "stmt_1b.png"
// 2015/05/26 15:50:44 Creating: "stmt_2a.dot"
// 2015/05/26 15:50:44 Creating: "stmt_2a.png"
// 2015/05/26 15:50:44 Creating: "stmt_2b.dot"
// 2015/05/26 15:50:44 Creating: "stmt_2b.png"
// 2015/05/26 15:50:44 Creating: "stmt_3a.dot"
// 2015/05/26 15:50:44 Creating: "stmt_3a.png"
// 2015/05/26 15:50:45 Creating: "stmt_3b.dot"
// 2015/05/26 15:50:45 Creating: "stmt_3b.png"
// ...
// 2015/05/26 15:50:54 Creating: "stmt_21a.dot"
// 2015/05/26 15:50:54 Creating: "stmt_21a.png"
// 2015/05/26 15:50:54 Creating: "stmt_21b.dot"
// 2015/05/26 15:50:54 Creating: "stmt_21b.png"
```

OUTPUT:
* [stmt.gif]: Intermediate CFGs created during the control flow recovery of [stmt.dot].

![Intermediate CFGs for stmt.dot][stmt.gif]

[stmt.gif]: https://raw.githubusercontent.com/decomp/restructure/master/examples/stmt.gif

## Dependencies

* [Go version 1.4](https://golang.org/doc/go1.4) (see [issue #1](https://github.com/decomp/restructure/issues/1) for details).

## Public domain

The source code and any original content of this repository is hereby released into the [public domain].

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/
