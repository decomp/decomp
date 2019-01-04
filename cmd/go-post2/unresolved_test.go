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
}
