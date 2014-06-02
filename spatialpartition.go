package go4game

import (
	//"log"
	"math"
)

type CheckSPObjListFn func(SPObjList) bool

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

func MakeSpatialPartition(w *World) *SpatialPartition {
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

func (p *SpatialPartition) ApplyParts27Fn(fn CheckSPObjListFn, pos Vector3D) bool {
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

func (p *SpatialPartition) makeRange2(c float64, r float64, min float64, max float64, n int) []int {
	if n-1 >= 0 && c-r <= min {
		return []int{n, n - 1}
	} else if n+1 < p.PartCount && c+r >= max {
		return []int{n, n + 1}
	} else {
		return []int{n}
	}
}

func (p *SpatialPartition) QueryByLen(fn CheckSPObjListFn, pos Vector3D, r float64) bool {
	ppos := p.Pos2PartPos(pos)
	partcube := p.GetPartCube(ppos)
	xr := p.makeRange2(pos[0], r, partcube.Min[0], partcube.Max[0], ppos[0])
	yr := p.makeRange2(pos[1], r, partcube.Min[1], partcube.Max[1], ppos[1])
	zr := p.makeRange2(pos[2], r, partcube.Min[2], partcube.Max[2], ppos[2])

	for _, x := range xr {
		for _, y := range yr {
			for _, z := range zr {
				cp := p.Parts[x][y][z]
				cpcube := p.GetPartCube([3]int{x, y, z})
				if len(cp) == 0 || !cpcube.IsContact(pos, r) {
					continue
				}
				if fn(cp) {
					return true
				}

			}
		}
	}
	return false
}
