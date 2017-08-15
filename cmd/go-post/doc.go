// The go-post tool post-processes Go source code to make it more idiomatic
// (*.go -> *.go).
//
// The input of go-post is unpolished Go source code and the output is more
// idiomatic Go source code.
package main

//go:generate usagen -o z_usage.go go-post

// Note, the usagen tool is located at github.com/mewmew/playground/cmd/usagen
