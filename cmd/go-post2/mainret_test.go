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
	os.Exit(int(42))
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
		os.Exit(int(i))
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
	// i=5,
	{
		Name: "mainret.5",
		In: `package main

func f() {
	return 42
}
`,
		Out: `package main

func f() {
	return 42
}
`,
	},
	// i=6,
	{
		Name: "mainret.6",
		In: `package p

func main() {
	return 42
}
`,
		Out: `package p

func main() {
	return 42
}
`,
	},
	// i=7,
	{
		Name: "mainret.7",
		In: `package main

func main(argc int32, argv **int8) {
	return 0
}
`,
		Out: `package main

import (
	"os"
	"unsafe"
)

func main() {
	argc := int32(len(os.Args))
	argv := (**int8)(unsafe.Pointer(&os.Args[0]))

}
`,
	},
}
