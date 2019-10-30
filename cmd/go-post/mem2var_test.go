package main

func init() {
	addTestCases(mem2varTests, mem2var)
}

var mem2varTests = []testCase{
	// i=0,
	{
		Name: "mem2var.0",
		In: `package main

func main() {
	var _12 int32
	_7 := new(int32)
	*_7 = 0
	_12 = *_7
}
`,
		Out: `package main

func main() {
	var _12 int32
	var _7 int32
	_7 = 0
	_12 = _7
}
`,
	},
}
