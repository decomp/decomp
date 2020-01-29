# The decomp.org project

[![Build Status](https://travis-ci.org/decomp/decomp.svg?branch=master)](https://travis-ci.org/decomp/decomp)
[![Coverage Status](https://coveralls.io/repos/github/decomp/decomp/badge.svg?branch=master)](https://coveralls.io/github/decomp/decomp?branch=master)
[![GoDoc](https://godoc.org/github.com/decomp/decomp?status.svg)](https://godoc.org/github.com/decomp/decomp)

The aim of this project is to implement a decompilation pipeline composed of independent components interacting through well-defined interfaces, as further described in the [design documents](https://github.com/decomp/doc) of the project.

## Installation

```bash
go get github.com/decomp/decomp/...
```

## Usage

See example usage at [examples/demo](examples/demo), and [this comment](https://github.com/decomp/decomp/issues/218#issuecomment-548506064) for further details.

## Decompilation pipeline

From a high-level perspective, the components of the decompilation pipeline are conceptually grouped into three modules. Firstly, the [front-end](#front-end) translates a source language (e.g. x86 assembly) into [LLVM IR](http://llvm.org/docs/LangRef.html); a platform-independent low-level intermediate representation. Secondly, the [middle-end](#middle-end) structures the LLVM IR by identifying high-level control flow primitives (e.g. pre-test loops, 2-way conditionals). Lastly, the [back-end](#back-end) translates the structured LLVM IR into a high-level target programming language (e.g. [Go](https://golang.org/)).

The following poster summarizes the current capabilities of the decompilation pipeline, using a composition of independent components to translate LLVM IR to Go.

[![Poster: Compositional Decompilation](https://raw.githubusercontent.com/decomp/doc/master/poster/poster.png)](https://raw.githubusercontent.com/decomp/doc/master/poster/poster.pdf)

### Front-end

Translate machine code (e.g. x86 assembly) to LLVM IR.

[Third-party front-end components](front-end.md).

### Middle-end

Perform control flow analysis on the LLVM IR to identify high-level control flow primitives (e.g. pre-test loops).

#### ll2dot

https://godoc.org/github.com/decomp/decomp/cmd/ll2dot

Control flow graph generation tool.

> Generate control flow graphs from LLVM IR assembly (*.ll -> *.dot).

#### restructure

https://godoc.org/github.com/decomp/decomp/cmd/restructure

Control flow recovery tool.

> Recover control flow primitives from control flow graphs (*.dot -> *.json).

### Back-end

Translate structured LLVM IR to a high-level target language (e.g. Go).

#### ll2go

https://godoc.org/github.com/decomp/decomp/cmd/ll2go

Go code generation tool.

> Decompile LLVM IR assembly to Go source code (*.ll -> *.go).

#### go-post

https://godoc.org/github.com/decomp/decomp/cmd/go-post

Go post-processing tool.

> Post-process Go source code to make it more idiomatic (*.go -> *.go).

## Release history

### Version 0.2 (2018-01-30)

Primary focus of version 0.2: *project-wide compilation speed*.

*Developing decompilation components should be fun.*

There seem to be an inverse correlation between depending on a huge C++ library and having fun developing decompilation components.

Version 0.2 of the decompilation pipeline strives to resolve this issue by leveraging an [LLVM IR library](https://github.com/llir/llvm) written in pure Go. Prior to this release, project-wide compilation could take several hours to complete. Now, they complete in less than 1 minute -- the established *hard limit* for all future releases.

### Version 0.1 (2015-04-21)

Initial release.

Primary focus of version 0.1: *compositional decompilation*.

*Decompilers should be composable and open source.*

A decompilation pipeline should be composed of individual components, each with a single purpose and well-defined input and output.

Version 0.1 of the decomp project explores the feasibility of composing a decompilation pipeline from independent components, and the potential of exposing those components to the end-user.

For further background, refer to the [Compositional Decompilation using LLVM IR](https://github.com/decomp/doc/raw/master/report/compositional_decompilation/compositional_decompilation.pdf) design document.

## Roadmap

### Version 0.3 (to be released)

Primary focus of version 0.3: *type-aware binary lifting*.

*Decompilers rely on high-quality binary lifting.*

The quality of the output IR of the binary lifting front-end fundamentally determines the quality of the output of the decompilation pipeline.

Version 0.3 aims to improve the quality of the output LLVM IR by implementing a type-aware binary lifting front-end.

### Version 0.4 (to be released)

Primary focus of version 0.4: *control flow analysis*.

*Decompilers should recover high-level control flow primitives.*

One of the primary differences between low-level assembly and high-level source code is the use of high-level control flow primitives; e.g. 1-way, 2-way and n-way conditionals (`if`, `if-else` and `switch`), pre- and post-test loops (`while` and `do-while`).

Version 0.4 seeks to recover high-level control flow primitives using robust control flow analysis algorithms.

### Version 0.5 (to be released)

Primary focus of version 0.5: *fault tolerance*.

*Decompilers should be robust.*

Decompilation components should respond well to unexpected states and incomplete analysis.

Version 0.5 focuses on stability, and seeks to stress test the decompilation pipeline using semi-real world software (see the [challenge issue series](https://github.com/decomp/decomp/labels/challenge)).

### Version 0.6 (to be released)

Primary focus of version 0.6: *data flow analysis*.

### Version 0.7 (to be released)

Primary focus of version 0.7: *type analysis*.
