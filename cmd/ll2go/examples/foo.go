// +build ignore

package main

func main() {
	i = 0
	x = 0
	for i < 10 {
		_4 := x < 100
		x = x
		if _4 {
			_6 := 3 * i
			_7 := x + _6
			x = _7
		}
		_10 := i + 1
		i = _10
		x = x
	}
	return x
}
