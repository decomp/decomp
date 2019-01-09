package main

func init() {
	addTestCases(killposTests, killpos)
}

var killposTests = []testCase{
	// i=0,
	{
		Name: "killpos.0",
		In: `package main

import (
	"os"
	"unsafe"
)

func main() {
	_0 := int32(len(os.Args))
	_1 := (**int8)(unsafe.Pointer(&os.Args[0]))
	_3 := new(int32)
	_4 := new(int32)
	_5 := new(**int8)
	_6 := new(int32)
	_7 := new(int32)
	*_3 = 0
	*_4 = _0
	*_5 = _1
	*_7 = 0
	*_6 = 0

	for *_6 <
		10 {

		if *_7 <
			100 {

			*_7 = *_7 +
				3*
					*_6

		}

		*_6 = *_6 +
			1

	}
	os.Exit(int(*_7))

}
`,
		Out: `package main

import (
	"os"
	"unsafe"
)

func main() {
	_0 := int32(len(os.Args))
	_1 := (**int8)(unsafe.Pointer(&os.Args[0]))
	_3 := new(int32)
	_4 := new(int32)
	_5 := new(**int8)
	_6 := new(int32)
	_7 := new(int32)
	*_3 = 0
	*_4 = _0
	*_5 = _1
	*_7 = 0
	*_6 = 0
	for *_6 < 10 {
		if *_7 < 100 {
			*_7 = *_7 + 3**_6
		}
		*_6 = *_6 + 1
	}

	os.Exit(int(*_7))
}
`,
	},
}
