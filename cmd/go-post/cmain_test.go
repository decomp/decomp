package main

func init() {
	addTestCases(cmainTests, cmain)
}

var cmainTests = []testCase{
	// i=0,
	{
		Name: "cmain.0",
		In: `package main

func main(_0 int32, _1 **int8) int32 {
	return 42
}
`,
		Out: `package main

import "os"

func c_main(_0 int32, _1 **int8) int32 {
	return 42
}
func main() {
	ret := int(c_main(0, nil))
	os.Exit(ret)
}
`,
	},
}
