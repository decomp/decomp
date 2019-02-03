// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(assignbinopTests, assignbinop)
}

var assignbinopTests = []testCase{
	// i=0,
	{
		Name: "assignbinop.0",
		In: `package main

func main() {
	x = x - 1
}
`,
		Out: `package main

func main() {
	x--
}
`,
	},
	// i=1,
	{
		Name: "assignbinop.1",
		In: `package main

func main() {
	x = 1 + x
}
`,
		Out: `package main

func main() {
	x++
}
`,
	},
	// i=2,
	{
		Name: "assignbinop.2",
		In: `package main

func main() {
	x = x / (5 - 3)
}
`,
		Out: `package main

func main() {
	x /= (5 - 3)
}
`,
	},
	// i=3,
	{
		Name: "assignbinop.3",
		In: `package main

func main() {
	i := 0
	for i < 10 {
		i = i + 1
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
	// i=4,
	{
		Name: "assignbinop.3",
		In: `package main

func main() {
	i := 0
	for i < 10 {
		i = i + 2
	}
}
`,
		Out: `package main

func main() {
	i := 0
	for i < 10 {
		i += 2
	}
}
`,
	},
}
