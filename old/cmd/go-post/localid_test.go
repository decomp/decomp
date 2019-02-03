// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(localidTests, localid)
}

// TODO: Remember to add parenthesis to test cases when the ast.ParenExpr TODO
// in localid.go is fixed.

var localidTests = []testCase{
	// i=0,
	{
		Name: "localid.0",
		In: `package main

func main() {
	_0 := i < 10
	if _0 {
		i = 20
	}
}
`,
		Out: `package main

func main() {

	if i < 10 {
		i = 20
	}
}
`,
	},
	// i=1,
	{
		Name: "localid.1",
		In: `package main

func main() {
	i := 0
	_0 = i < 10
	for _0 {
		_1 := i + 1
		i = _1
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
		Name: "localid.2",
		In: `package main

func main() {
	_0 := i + j
	_1 := x * y
	a := _0 + _1
}
`,
		Out: `package main

func main() {

	a := i + j +
		x*y

}
`,
	},
}
