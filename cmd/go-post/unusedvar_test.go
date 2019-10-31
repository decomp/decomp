package main

func init() {
	addTestCases(unusedvarTests, unusedvar)
}

var unusedvarTests = []testCase{
	// i=0,
	{
		Name: "unusedvar.0",
		In: `package p

func f() {
	var x int
}
`,
		Out: `package p

func f() {

}
`,
	},
	/*
	// i=1,
	{
		Name: "unusedvar.1",
		In: `package p

func g(int a) {
}

func f(int a) {
	var x int
	var y int
	x = a
	y = a
	g(y)
}
`,
		Out: `package p

func g(int a) {
}

func f(int a) {
	var y int
	y = a
	g(y)
}
`,
	},
	*/
}
