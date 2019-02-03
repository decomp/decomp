// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(deadassignTests, deadassign)
}

// TODO: Remember to add parenthesis to test cases when the ast.ParenExpr TODO
// in deadassign.go is fixed.

var deadassignTests = []testCase{
	// i=0,
	{
		Name: "deadassign.0",
		In: `package main

func main() {
	x = x
}
`,
		Out: `package main

func main() {

}
`,
	},
	// i=1,
	{
		Name: "deadassign.1",
		In: `package main

func main() {
	sum := 0
	i := 0
	for i < 10 {
		sum += i
		i++
		sum = sum
	}
}
`,
		Out: `package main

func main() {
	sum := 0
	i := 0
	for i < 10 {
		sum += i
		i++

	}
}
`,
	},
}
