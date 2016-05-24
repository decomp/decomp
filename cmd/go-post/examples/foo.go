package main

import "os"

func main() {
	x := 0
	for i := 0; i < 10; i++ {
		if x < 100 {
			x += 3 * i
		}
	}
	os.Exit(x)
}
