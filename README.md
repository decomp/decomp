# The decomp.org project

The aim of this project is to implement a decompilation pipeline composed of independent components interacting through well-defined interfaces, as further described in the [design documents](https://github.com/decomp/doc) of the project.

From a high-level perspective, the components of the decompilation pipeline are conceptually grouped into three modules. Firstly, the *front-end* translates a source language (e.g. x86 assembly) into [LLVM IR](http://llvm.org/docs/LangRef.html); a platform-independent low-level intermediate representation. Secondly, the *middle-end* structures the LLVM IR by identifying high-level control flow primitives (e.g. pre-test loops, 2-way conditionals). Lastly, the *back-end* translates the structured LLVM IR into a high-level target programming language (e.g. [Go](https://golang.org/)).

The following poster summarizes the current capabilities of the decompilation pipeline, using a composition of independent components to translate LLVM IR to Go.

[![Poster: Compositional Decompilation](https://raw.githubusercontent.com/decomp/doc/master/poster/poster.png)](https://raw.githubusercontent.com/decomp/doc/master/poster/poster.pdf)

## Front-end

Translate machine code (e.g. x86 assembly) to LLVM IR.

[Third-party front-end components](front-end.md).

## Middle-end

Perform control flow analysis on the LLVM IR to identify high-level control flow primitives (e.g. pre-test loops).

### ll2dot

https://godoc.org/github.com/decomp/decomp/cmd/ll2dot

Control flow graph generation tool.

> Generate control flow graphs from LLVM IR assembly (*.ll -> *.dot).

### restructure

https://godoc.org/github.com/decomp/decomp/cmd/restructure

Control flow recovery tool.

> Recover control flow primitives from control flow graphs (*.dot -> *.json).

## Back-end

Translate structured LLVM IR to a high-level target language (e.g. Go).

### ll2go

https://godoc.org/github.com/decomp/decomp/cmd/ll2go

Go code generation tool.

> Decompile LLVM IR assembly to Go source code (*.ll -> *.go).

### go-post

https://godoc.org/github.com/decomp/decomp/cmd/go-post

Go post-processing tool.

> Post-process Go source code to make it more idiomatic (*.go -> *.go).

## Public domain

The source code and any original content of this repository is hereby released into the [public domain].

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/
