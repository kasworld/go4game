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
	Parts           [][][]SPObjList
	MaxObjectRadius float64
}

func (p *SpatialPartition) GetPartPos(pos *Vector3D) [3]int {
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

func (p *SpatialPartition) touchBottom(iaxis int, pposn int, pos *Vector3D) bool {
	return pposn-1 >= 0 && p.PartSize[iaxis]*float64(pposn)+p.Min[iaxis]+p.MaxObjectRadius*2 >= pos[iaxis]
}
func (p *SpatialPartition) touchTop(iaxis int, pposn int, pos *Vector3D) bool {
	return pposn+1 < p.PartCount && p.PartSize[iaxis]*float64(pposn+1)+p.Min[iaxis]-p.MaxObjectRadius*2 <= pos[iaxis]
}

func (p *SpatialPartition) makeRange(m *GameObject, ppos [3]int, iaxis int) []int {
	if p.touchBottom(iaxis, ppos[iaxis], &m.PosVector) {
		return []int{ppos[iaxis], ppos[iaxis] - 1}
	} else if p.touchTop(iaxis, ppos[iaxis], &m.PosVector) {
		return []int{ppos[iaxis], ppos[iaxis] + 1}
	} else {
		return []int{ppos[iaxis]}
	}
}

func (p *SpatialPartition) ApplyCollisionAction3(fn CollisionActionFn, m *GameObject) bool {
	ppos := p.GetPartPos(&m.PosVector)

	xr := p.makeRange(m, ppos, 0)
	yr := p.makeRange(m, ppos, 1)
	zr := p.makeRange(m, ppos, 2)
	//log.Printf("%v %v %v ", xr, yr, zr)
	for _, i := range xr {
		for _, j := range yr {
			for _, k := range zr {
				if p.ApplyPartFn(fn, m, [...]int{i, j, k}) {
					return true
				}
			}
		}
	}
	return false
}

// min <= v < max
func getCheckRange(v int, min int, max int) []int {
	if v <= min {
		return []int{min, min + 1}
	} else if v+1 >= max {
		return []int{max - 1, max - 2}
	} else {
		return []int{v, v - 1, v + 1}
	}
}
func (p *SpatialPartition) ApplyCollisionAction1(fn CollisionActionFn, m *GameObject) bool {
	ppos := p.GetPartPos(&m.PosVector)
	xr := getCheckRange(ppos[0], 0, p.PartCount)
	yr := getCheckRange(ppos[1], 0, p.PartCount)
	zr := getCheckRange(ppos[2], 0, p.PartCount)
	for _, i := range xr {
		for _, j := range yr {
			for _, k := range zr {
				if p.ApplyPartFn(fn, m, [...]int{i, j, k}) {
					return true
				}
			}
		}
	}
	return false
}

type CollisionActionFn func(s *SPObj, m *GameObject) bool

func IsCollision(s *SPObj, target *GameObject) bool {
	teamrule := s.TeamID != target.PTeam.ID
	checklen := s.PosVector.LenTo(&target.PosVector) <= (s.CollisionRadius + target.CollisionRadius)
	return (teamrule) && (checklen)
}
func (p *SpatialPartition) ApplyPartFn(fn CollisionActionFn, m *GameObject, ppos [3]int) bool {
	for _, v := range p.Parts[ppos[0]][ppos[1]][ppos[2]] {
		if fn(v, m) {
			return true
		}
	}
	return false
}

func (p *SpatialPartition) ApplyCollisionAction2(fn CollisionActionFn, m *GameObject) bool {
	ppos := p.GetPartPos(&m.PosVector)
	for i := ppos[0] - 1; i <= ppos[0]+1; i++ {
		if i < 0 || i >= p.PartCount {
			continue
		}
		for j := ppos[1] - 1; j <= ppos[1]+1; j++ {
			if j < 0 || j >= p.PartCount {
				continue
			}
			for k := ppos[2] - 1; k <= ppos[2]+1; k++ {
				if k < 0 || k >= p.PartCount {
					continue
				}
				if p.ApplyPartFn(fn, m, [...]int{i, j, k}) {
					return true
				}
			}
		}
	}
	return false
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
	if rtn.PartCount < 2 {
		rtn.PartCount = 2
	}
	rtn.PartSize = *rtn.Size.Idiv(float64(rtn.PartCount))

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
				partPos := rtn.GetPartPos(&obj.PosVector)
				rtn.AddPartPos(partPos, NewSPObj(obj))
			}
		}
	}
	return &rtn
}
