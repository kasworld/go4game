package go4game

import (
	"math/rand"
)

const (
	TeamTypeNil = iota
	TeamTypePlayer
	TeamTypeAI
	TeamTypeObserver
	TeamTypeTerrain
	TeamTypeFood
	TeamTypeDeco
	TeamTypeEnd
)

func (t *Team) addFood() {
	hr := GameConst.WorldCube
	for i := 0; i < 50; i++ {
		pos := hr.RandVector()
		o := NewGameObject(t.ID).MakeFoodObj(pos)
		t.addObject(o)
	}
}

func (t *Team) addTerrain() {
	w := GameConst.MaxObjectRadius * 3
	hr := GameConst.WorldCube2
	for i := hr.Min[0]; i <= hr.Max[0]; i += w {
		for j := hr.Min[1]; j <= hr.Max[1]; j += w {
			for k := hr.Min[2]; k <= hr.Max[2]; k += w {
				if rand.Float64() < 0.5 {
					continue
				}
				pos := Vector3D{i, j, k}
				o := NewGameObject(t.ID).MakeHardObj(pos)
				t.addObject(o)
			}
		}
	}
}

func (t *Team) addMaze() {
	w := GameConst.MaxObjectRadius
	hr := GameConst.WorldCube2
	for i := hr.Min[0]; i < hr.Max[0]; i += w {
		j := 0.0
		for k := hr.Min[2]; k < hr.Max[2]; k += w {
			if rand.Float64() < 0.8 {
				continue
			}
			pos := Vector3D{i, j, k}
			o := NewGameObject(t.ID).MakeHardObj(pos)
			t.addObject(o)
		}
	}
}

func (t *Team) addRevolutionDeco() {
	avt := GameConst.WorldCube.RandVector().Idiv(10)
	mvvt := GameConst.WorldCube.RandVector().Idiv(10)
	for i := 0; i < 50; i++ {
		o := NewGameObject(t.ID).MakeRevolutionDecoObj()

		o.accelVector = avt //.NormalizedTo(float64(i * 10+1))
		o.MoveVector = mvvt.NormalizedTo(float64(i*16 + 1))
		t.addObject(o)
	}
}
