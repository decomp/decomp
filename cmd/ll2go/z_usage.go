/*
Usage: ll2go [OPTION]... FILE...
Decompile LLVM IR assembly files to Go source code (e.g. *.ll -> *.go).

Flags:
  -f    Force overwrite existing Go source code.
  -funcs string
        Comma separated list of functions to decompile (e.g. "foo,bar").
  -pkgname string
        Package name.
  -q    Suppress non-error messages.
  -v    Enable verbose output.
*/
package main
