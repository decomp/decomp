// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(deadlabelTests, deadlabel)
}

var deadlabelTests = []testCase{
	// i=0,
	{
		Name: "deadlabel.0",
		In: `package main

func main() {
foo:
	i := 0
bar:
	for i < 10 {
	baz:
		i++
	}
}
`,
		Out: `package main

func main() {

	i := 0

	for i < 10 {

		i++
	}
}
`,
	},
	// i=1,
	{
		Name: "deadlabel.1",
		In: `package main

func main() {
main:
	sum := 0
foo:
	i := 0
loop:
	for i < 10 {
	bar:
		sum += i
	baz:
		if sum > 10 {
		qux:
			break loop
		}
	gob:
		i++
	}
}
`,
		Out: `package main

func main() {

	sum := 0

	i := 0
loop:
	for i < 10 {

		sum += i

		if sum > 10 {

			break loop
		}

		i++
	}
}
`,
	},
}
