package go4game

import (
	//"log"
	"math"
)

type SPObj struct {
	ID         int64
	TeamID     int64
	PosVector  Vector3D
	MoveVector Vector3D
	ObjType    GameObjectType
}

func NewSPObj(o *GameObject) *SPObj {
	if o == nil {
		return nil
	}
	return &SPObj{
		ID:         o.ID,
		TeamID:     o.TeamID,
		PosVector:  o.PosVector,
		MoveVector: o.MoveVector,
		ObjType:    o.ObjType,
	}
}

type SPObjList []*SPObj

type SpatialPartition struct {
	WorldCube HyperRect
	Size      Vector3D

	PartCount int
	PartSize  Vector3D
	PartLen   float64
	PartMins  []Vector3D

	Parts       [][][]SPObjList
	ObjectCount int
}

func (p *SpatialPartition) AddPartPos(pos [3]int, obj *SPObj) {
	p.Parts[pos[0]][pos[1]][pos[2]] = append(p.Parts[pos[0]][pos[1]][pos[2]], obj)
}

func (w *World) MakeSpatialPartition() *SpatialPartition {
	rtn := SpatialPartition{
		WorldCube: GameConst.WorldCube,
		Size:      GameConst.WorldCube.SizeVector(),
	}
	objcount := 0
	for _, t := range w.Teams {
		objcount += len(t.GameObjs)
	}
	rtn.ObjectCount = objcount

	rtn.PartCount = int(math.Pow(float64(objcount), 1.0/3.0))
	if rtn.PartCount < 3 {
		rtn.PartCount = 3
	}
	rtn.PartSize = rtn.Size.Idiv(float64(rtn.PartCount))
	rtn.PartLen = rtn.PartSize.Abs()
	rtn.PartMins = make([]Vector3D, rtn.PartCount+1)
	rtn.PartMins[0] = rtn.WorldCube.Min
	for i := 1; i < rtn.PartCount; i++ {
		rtn.PartMins[i] = rtn.PartMins[i-1].Add(rtn.PartSize)
	}
	rtn.PartMins[rtn.PartCount] = rtn.WorldCube.Max

	rtn.Parts = make([][][]SPObjList, rtn.PartCount)
	for i := 0; i < rtn.PartCount; i++ {
		rtn.Parts[i] = make([][]SPObjList, rtn.PartCount)
		for j := 0; j < rtn.PartCount; j++ {
			rtn.Parts[i][j] = make([]SPObjList, rtn.PartCount)
		}
	}

	for _, t := range w.Teams {
		for _, obj := range t.GameObjs {
			if obj != nil && obj.ObjType != 0 {
				partPos := rtn.Pos2PartPos(obj.PosVector)
				rtn.AddPartPos(partPos, NewSPObj(obj))
			}
		}
	}
	return &rtn
}

func (p *SpatialPartition) Pos2PartPos(pos Vector3D) [3]int {
	nompos := pos.Sub(p.WorldCube.Min)
	rtn := [3]int{0, 0, 0}

	for i, v := range nompos {
		rtn[i] = int(v / p.PartSize[i])
		if rtn[i] >= p.PartCount {
			rtn[i] = p.PartCount - 1
		} else if rtn[i] < 0 {
			rtn[i] = 0
		}
	}
	return rtn
}

func (p *SpatialPartition) GetPartCube(ppos [3]int) *HyperRect {
	return &HyperRect{
		Min: Vector3D{p.PartMins[ppos[0]][0], p.PartMins[ppos[1]][1], p.PartMins[ppos[2]][2]},
		Max: Vector3D{p.PartMins[ppos[0]+1][0], p.PartMins[ppos[1]+1][1], p.PartMins[ppos[2]+1][2]},
	}
}

func (p *SpatialPartition) IsContactTo(c Vector3D, ppos [3]int, plenrsqd float64) bool {
	var sum float64
	for i, v := range c {
		d := p.PartMins[ppos[i]][i] + p.PartSize[i]/2 - v
		sum += d * d
	}
	return plenrsqd >= sum
	// return plenrsqd >= c.Sqd(Vector3D{
	// 	p.PartMins[x][0] + p.PartSize[0]/2,
	// 	p.PartMins[y][1] + p.PartSize[1]/2,
	// 	p.PartMins[z][2] + p.PartSize[2]/2})
	// pMin := Vector3D{p.PartMins[x][0], p.PartMins[y][1], p.PartMins[z][2]}
	// pCenter := pMin.Add(p.PartSize.Idiv(2))
	// return plenrsqd >= pCenter.Sqd(c)
}

func (p *SpatialPartition) getRangeStart(n int) int {
	if n <= 1 {
		return 0
	} else if n >= p.PartCount-2 {
		return p.PartCount - 3
	} else {
		return n - 1
	}
}

func (p *SpatialPartition) getPartStart27(pos Vector3D) (x, y, z int) {
	ppos := p.Pos2PartPos(pos)
	x = p.getRangeStart(ppos[0])
	y = p.getRangeStart(ppos[1])
	z = p.getRangeStart(ppos[2])
	return
}

func (p *SpatialPartition) ApplyParts27Fn(fn func(SPObjList) bool, pos Vector3D) bool {
	i, j, k := p.getPartStart27(pos)
	for x := i; x < i+3; x++ {
		for y := j; y < j+3; y++ {
			for z := k; z < k+3; z++ {
				if fn(p.Parts[x][y][z]) {
					return true
				}
			}
		}
	}
	return false
}

func (p *SpatialPartition) ApplyParts27Fn2(fn func(SPObjList, [3]int) bool, pos Vector3D) bool {
	i, j, k := p.getPartStart27(pos)
	for x := i; x < i+3; x++ {
		for y := j; y < j+3; y++ {
			for z := k; z < k+3; z++ {
				if fn(p.Parts[x][y][z], [3]int{x, y, z}) {
					return true
				}
			}
		}
	}
	return false
}
