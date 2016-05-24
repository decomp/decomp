## WIP

This project is a *work in progress*. The implementation is *incomplete* and subject to change. The documentation can be inaccurate.

# ll2go

[![GoDoc](https://godoc.org/decomp.org/x/cmd/ll2go?status.svg)](https://godoc.org/decomp.org/x/cmd/ll2go)

`ll2go` is a tool which decompiles LLVM IR assembly files to Go source code (e.g. *.ll -> *.go).

Please note that `ll2go` is meant to be used in combination with [go-post] which post-processes the Go source code to make it more idiomatic.

[go-post]: https://github.com/decomp/go-post

## Installation

* Install the [dependencies](https://github.com/decomp/ll2go#dependencies) before running go-get.

```shell
go get decomp.org/x/cmd/ll2go
```

## Usage

```
Usage: ll2go [OPTION]... FILE...

Flags:
  -f  Force overwrite existing Go source code.
  -funcs string
      Comma separated list of functions to decompile (e.g. "foo,bar").
  -pkgname string
      Package name.
  -q  Suppress non-error messages.
  -v  Enable verbose output.
```

## Examples

```bash
$ ll2go foo.ll
```

INPUT: [foo.ll](examples/foo.ll)

```llvm
define i32 @main(i32 %argc, i8** %argv) {
  br label %1

; <label>:1                                       ; preds = %9, %0
  %i.0 = phi i32 [ 0, %0 ], [ %10, %9 ]
  %x.0 = phi i32 [ 0, %0 ], [ %x.1, %9 ]
  %2 = icmp slt i32 %i.0, 10
  br i1 %2, label %3, label %11

; <label>:3                                       ; preds = %1
  %4 = icmp slt i32 %x.0, 100
  br i1 %4, label %5, label %8

; <label>:5                                       ; preds = %3
  %6 = mul nsw i32 3, %i.0
  %7 = add nsw i32 %x.0, %6
  br label %8

; <label>:8                                       ; preds = %5, %3
  %x.1 = phi i32 [ %7, %5 ], [ %x.0, %3 ]
  br label %9

; <label>:9                                       ; preds = %8
  %10 = add nsw i32 %i.0, 1
  br label %1

; <label>:11                                      ; preds = %1
  ret i32 %x.0
}
```

OUTPUT: [foo.go](examples/foo.go)

Unpolished Go output; polish using [go-post].

```go
package main

func main() {
    i = 0
    x = 0
    for i < 10 {
        _4 := x < 100
        x = x
        if _4 {
            _6 := 3 * i
            _7 := x + _6
            x = _7
        }
        _10 := i + 1
        i = _10
        x = x
    }
    return x
}
```

## Dependencies

* [llvm.org/llvm/bindings/go/llvm](https://godoc.org/llvm.org/llvm/bindings/go/llvm) with [unnamed.patch](https://raw.githubusercontent.com/decomp/ll2dot/master/unnamed.patch)
* `llvm-as` from [LLVM](http://llvm.org/)
* `dot` from [Graphviz](http://www.graphviz.org/)
* [ll2dot](https://decomp.org/x/cmd/ll2dot)

## Public domain

The source code and any original content of this repository is hereby released into the [public domain].

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/
