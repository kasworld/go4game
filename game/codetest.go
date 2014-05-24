package main

import (
	"fmt"
)

type int3 [3]int

func (i int3) mod(v int, ax int) int3 {
	i[ax] += v
	return i
}

func permute(pos int3, up int3, down int3, ax int) {
	if up[ax] < 0 || down[ax] > 0 {
		return
	}
	if ax == 2 {
		fmt.Printf("%#v", pos)
		fmt.Printf("%#v", pos.mod(1, ax))
		fmt.Printf("%#v\n", pos.mod(-1, ax))
	} else {
		permute(pos, up, down, ax+1)
		permute(pos.mod(1, ax), up.mod(-1, ax), down, ax+1)
		permute(pos.mod(-1, ax), up, down.mod(1, ax), ax+1)
	}
}

func main() {
	permute(int3{0, 0, 0}, int3{1, 1, 1}, int3{-1, -1, -1}, 0)
}

var p27 = [27][3]int{
	{0, 0, 0}, {0, 0, 1}, {0, 0, -1},
	{0, 1, 0}, {0, 1, 1}, {0, 1, -1},
	{0, -1, 0}, {0, -1, 1}, {0, -1, -1},
	{1, 0, 0}, {1, 0, 1}, {1, 0, -1},
	{1, 1, 0}, {1, 1, 1}, {1, 1, -1},
	{1, -1, 0}, {1, -1, 1}, {1, -1, -1},
	{-1, 0, 0}, {-1, 0, 1}, {-1, 0, -1},
	{-1, 1, 0}, {-1, 1, 1}, {-1, 1, -1},
	{-1, -1, 0}, {-1, -1, 1}, {-1, -1, -1},
}
