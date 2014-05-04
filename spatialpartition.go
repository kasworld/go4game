package go4game

import (
	"log"
	"math"
)

type SPObj struct {
	ID              int
	TeamID          int
	PosVector       Vector3D
	MoveVector      Vector3D
	CollisionRadius float64
	ObjType         GameObjectType
}

func NewSPObj(o *GameObject) *SPObj {
	return &SPObj{
		ID:              o.ID,
		TeamID:          o.PTeam.ID,
		PosVector:       o.PosVector,
		MoveVector:      o.MoveVector,
		CollisionRadius: o.CollisionRadius,
		ObjType:         o.ObjType,
	}
}

type SPObjList []*SPObj

type SpatialPartition struct {
	Min             Vector3D
	Max             Vector3D
	Size            Vector3D
	PartCount       int
	PartSize        Vector3D
	PartMins        []Vector3D
	Parts           [][][]SPObjList
	MaxObjectRadius float64
}

func (p *SpatialPartition) AddPartPos(pos [3]int, obj *SPObj) {
	p.Parts[pos[0]][pos[1]][pos[2]] = append(p.Parts[pos[0]][pos[1]][pos[2]], obj)
}

func (w *World) MakeSpatialPartition() *SpatialPartition {
	rtn := SpatialPartition{
		Min:             w.MinPos,
		Max:             w.MaxPos,
		Size:            *w.MaxPos.Sub(&w.MinPos),
		MaxObjectRadius: w.MaxObjectRadius,
	}
	objcount := 0
	for _, t := range w.Teams {
		objcount += len(t.GameObjs)
	}

	rtn.PartCount = int(math.Pow(float64(objcount), 1.0/3.0))
	if rtn.PartCount < 3 {
		rtn.PartCount = 3
	}
	rtn.PartSize = *rtn.Size.Idiv(float64(rtn.PartCount))
	rtn.PartMins = make([]Vector3D, rtn.PartCount+1)
	for i := 0; i < rtn.PartCount; i++ {
		rtn.PartMins[i] = *rtn.Min.Add(&Vector3D{
			float64(i) * rtn.PartSize[0],
			float64(i) * rtn.PartSize[1],
			float64(i) * rtn.PartSize[2]})
	}
	rtn.PartMins[rtn.PartCount] = rtn.Max

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
	nompos := pos.Sub(&p.Min)
	rtn := [3]int{0, 0, 0}

	for i, v := range nompos {
		rtn[i] = int(v / p.PartSize[i])
		if rtn[i] >= p.PartCount {
			rtn[i] = p.PartCount - 1
			//log.Printf("invalid pos %v %v", v, rtn[i])
		}
		if rtn[i] < 0 {
			rtn[i] = 0
			log.Printf("invalid pos %v %v", v, rtn[i])
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

func (p *SpatialPartition) makeRange2(c float64, r float64, min float64, max float64, n int) []int {
	if n-1 >= 0 && c-r*2 <= min {
		return []int{n, n - 1}
	} else if n+1 < p.PartCount && c+r*2 >= max {
		return []int{n, n + 1}
	} else {
		return []int{n}
	}
}

type PartsFn func(s SPObjList) bool

func (p *SpatialPartition) ApplyPartsFn(fn PartsFn, pos Vector3D, r float64) bool {
	ppos := p.Pos2PartPos(pos)
	partcube := p.GetPartCube(ppos)

	xr := p.makeRange2(pos[0], r, partcube.Min[0], partcube.Max[0], ppos[0])
	yr := p.makeRange2(pos[1], r, partcube.Min[1], partcube.Max[1], ppos[1])
	zr := p.makeRange2(pos[2], r, partcube.Min[2], partcube.Max[2], ppos[2])
	//log.Printf("%v %v %v ", xr, yr, zr)
	for _, i := range xr {
		for _, j := range yr {
			for _, k := range zr {
				if fn(p.Parts[i][j][k]) {
					return true
				}
			}
		}
	}
	return false
}

func (p *SpatialPartition) makeRange3(n int) []int {
	if n <= 1 {
		return []int{0, 1, 2}
	} else if n >= p.PartCount-2 {
		return []int{p.PartCount - 1, p.PartCount - 2, p.PartCount - 3}
	} else {
		return []int{n - 1, n, n + 1}
	}
}
func (p *SpatialPartition) ApplyParts27Fn(fn PartsFn, pos Vector3D, r float64) bool {
	ppos := p.Pos2PartPos(pos)
	xr := p.makeRange3(ppos[0])
	yr := p.makeRange3(ppos[1])
	zr := p.makeRange3(ppos[2])
	//log.Printf("%v %v %v ", xr, yr, zr)
	for _, i := range xr {
		for _, j := range yr {
			for _, k := range zr {
				if fn(p.Parts[i][j][k]) {
					return true
				}
			}
		}
	}
	return false
}
