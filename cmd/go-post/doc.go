// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate usagen go-post

/*
go-post post-processes Go source code to make it more idiomatic.

Without an explicit path, go-post reads standard input and writes the
result to standard output.

If the named path is a file, go-post rewrites the named files in place.
If the named path is a directory, go-post rewrites all .go files in that
directory tree.  When go-post rewrites a file, it prints a line to standard
error giving the name of the file and the rewrite applied.

If the -diff flag is set, no files are rewritten. Instead go-post prints
the differences a rewrite would introduce.

The -r flag restricts the set of rewrites considered to those in the
named list.  By default go-post considers all known rewrites.  go-post's
rewrites are idempotent, so that it is safe to apply go-post to updated
or partially updated code even without using the -r flag.

go-post prints the full list of fixes it can apply in its help output;
to see them, run go-post -help.

go-post does not make backup copies of the files that it edits.
Instead, use a version control system's ``diff'' functionality to inspect
the changes that go-post makes before committing them.
*/
package main
