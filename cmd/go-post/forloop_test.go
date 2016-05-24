// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(forloopTests, forloop)
}

var forloopTests = []testCase{
	// i=0,
	{
		Name: "forloop.0",
		In: `package main

func main() {
	i := 0
	for i < 10 {
		i++
	}
}
`,
		Out: `package main

func main() {

	for i := 0; i < 10; i++ {

	}
}
`,
	},
	// i=1,
	{
		Name: "forloop.1",
		In: `package main

func main() {
	sum := 0
	i := 0
	for i < 10 {
		sum += i
		i++
	}
}
`,
		Out: `package main

func main() {
	sum := 0

	for i := 0; i < 10; i++ {
		sum += i

	}
}
`,
	},
	// i=2,
	{
		Name: "forloop.2",
		In: `package main

import "fmt"

func main() {
	xs := []int{1, 2, 3, 4, 5}
	i := len(xs) - 1
	for i >= 0 {
		fmt.Println(i)
		i--
	}
}
`,
		Out: `package main

import "fmt"

func main() {
	xs := []int{1, 2, 3, 4, 5}

	for i := len(xs) - 1; i >= 0; i-- {
		fmt.Println(i)

	}
}
`,
	},
}
