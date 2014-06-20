package main

import (
	"encoding/json"
	"fmt"
	"github.com/kasworld/go4game"
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

func test_permute() {
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

func testjson() {
	var jsonBlob1 = []byte(`
		{"Name": "Platypus"}
	`)
	var jsonBlob2 = []byte(`
		{"Order": "Platypus2"}
	`)
	type Animal struct {
		Name  string
		Order string
	}
	var animals Animal
	err := json.Unmarshal(jsonBlob1, &animals)
	if err != nil {
		fmt.Println("error1:", err)
	}
	fmt.Printf("1%+v\n", animals)

	err = json.Unmarshal(jsonBlob2, &animals)
	if err != nil {
		fmt.Println("error2:", err)
	}
	fmt.Printf("2%+v\n", animals)
}

func main() {
	v1 := go4game.Vector3D{0, 0, 0}
	v2 := go4game.Vector3D{1, -1, 1}
	d8 := v1.To8Direct(v2)
	fmt.Printf("%v %v %v\n", v1, v2, d8)
	hr := go4game.HyperRect{
		go4game.Vector3D{-10, -10, -10},
		go4game.Vector3D{10, 10, 10},
	}
	nhr := hr.MakeCubeBy8Driect(v1, d8)
	fmt.Printf("%v %v %v isin %v\n", hr, nhr, d8, v2.IsIn(nhr))

}
