// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(unresolvedTests, unresolved)
}

var unresolvedTests = []testCase{
	// i=0,
	{
		Name: "unresolved.0",
		In: `package main

func main() {
	i = 42
}
`,
		Out: `package main

func main() {
	i := 42
}
`,
	},
	// i=1,
	{
		Name: "unresolved.1",
		In: `package main

func main() {
	i = 0
	for i < 10 {
		i = i + 1
	}
}
`,
		Out: `package main

func main() {
	i := 0
	for i < 10 {
		i = i + 1
	}
}
`,
	},
	// i=2,
	{
		Name: "unresolved.2",
		In: `package main

func main() {
	j := 10
	i, j = 5, 5
}
`,
		Out: `package main

func main() {
	j := 10
	i, j := 5, 5
}
`,
	},
	// i=3,
	{
		Name: "unresolved.3",
		In: `package main

var i int

func main() {
	for i < 10 {
		i = i + 1
	}
}
`,
		Out: `package main

var i int

func main() {
	for i < 10 {
		i = i + 1
	}
}
`,
	},
	/*
		// TODO: Make unresolved idempotent and enable this test case. If anyone can
		// think of a clean way to do this, please let me know.
		// i=4,
		{
			Name: "unresolved.4",
			In: `package main

			func main() {
				if j = 0; true {
					j = 10
				}
				if j = 1; true {
					j = 20
				}
			}
			`,
			Out: `package main

			func main() {
				if j := 0; true {
					j = 10
				}
				if j = 1; true {
					j = 20
				}
			}
			`,
		},
	*/
}
