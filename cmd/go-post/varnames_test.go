package main

func init() {
	addTestCases(varnamesTests, varnames)
}

var varnamesTests = []testCase{
	// i=0,
	{
		Name: "varnames.0",
		In: `package main

func main() {
	var _12 int32
	var _7 int32
	_7 = 0
	_12 = _7
}
`,
		Out: `package main

func main() {
	var v_12 int32
	var v_7 int32
	v_7 = 0
	v_12 = v_7
}
`,
	},
}
