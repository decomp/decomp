//+build examples

package main

import "os"

func c_main(_0 int32, _1 **int8) int32 {
	v_7 := int32(0)
	for v_6 := int32(0); v_6 < 10; v_6++ {
		if v_7 < 100 {
			v_7 += 3 * v_6
		}
	}
	return v_7
}

func main() {
	ret := int(c_main(0, nil))
	os.Exit(ret)
}
