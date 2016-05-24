// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(mainretTests, mainret)
}

var mainretTests = []testCase{
	// i=0,
	{
		Name: "mainret.0",
		In: `package main

func main() {
	return 42
}
`,
		Out: `package main

import "os"

func main() {
	os.Exit(42)
}
`,
	},
	// i=1,
	{
		Name: "mainret.1",
		In: `package main

func main() {
	return
}
`,
		Out: `package main

func main() {

}
`,
	},
	// i=2,
	{
		Name: "mainret.2",
		In: `package main

func main() {
	i := 42
	if i >= 128 {
		return i
	}
}
`,
		Out: `package main

import "os"

func main() {
	i := 42
	if i >= 128 {
		os.Exit(i)
	}
}
`,
	},
	// i=3,
	{
		Name: "mainret.3",
		In: `package main

func main() {
	return 0
}
`,
		Out: `package main

func main() {

}
`,
	},
	// i=4,
	{
		Name: "mainret.4",
		In: `package main

func main() {
	if true {
		return 0
	}
	return 0
}
`,
		Out: `package main

func main() {
	if true {
		return
	}

}
`,
	},
}
