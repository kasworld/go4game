package go4game

import (
	"log"
	"math"
)

type SpatialPartition struct {
	Min, Max Vector3D
	PartSize int
	refs     [][][]GameObjectList
}

func (p *SpatialPartition) GetPartPos(pos *Vector3D) [3]int {
	nompos := pos.Sub(&p.Min)
	rtn := [3]int{0, 0, 0}

	for i, v := range nompos {
		l := p.Max[i] - p.Min[i]
		rtn[i] = int(v / l * float64(p.PartSize))
		if rtn[i] >= p.PartSize {
			rtn[i] = p.PartSize - 1
			//log.Printf("invalid pos %v %v", v, rtn[i])
		}
		if rtn[i] < 0 {
			rtn[i] = 0
			log.Printf("invalid pos %v %v", v, rtn[i])
		}
	}
	return rtn
}

func (p *SpatialPartition) AddPartPos(pos [3]int, obj *GameObject) {
	p.refs[pos[0]][pos[1]][pos[2]] = append(p.refs[pos[0]][pos[1]][pos[2]], obj)
}

func get3(v int, min int, max int) []int {
	rtn := make([]int, 0, 3)
	rtn = append(rtn, v)
	if v-1 >= min {
		rtn = append(rtn, v-1)
	}
	if v+1 < max {
		rtn = append(rtn, v+1)
	}
	return rtn
}

func (p *SpatialPartition) IsCollision(m *GameObject) bool {
	ppos := p.GetPartPos(&m.posVector)
	for i := ppos[0] - 1; i <= ppos[0]+1; i++ {
		if i < 0 || i >= p.PartSize {
			continue
		}
		for j := ppos[1] - 1; j <= ppos[1]+1; j++ {
			if j < 0 || j >= p.PartSize {
				continue
			}
			for k := ppos[2] - 1; k <= ppos[2]+1; k++ {
				if k < 0 || k >= p.PartSize {
					continue
				}
				for _, v := range p.refs[i][j][k] {
					if m.IsCollision(v) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (w *World) MakeSpatialPartition() *SpatialPartition {
	rtn := SpatialPartition{
		Min: w.MinPos,
		Max: w.MaxPos,
	}
	objcount := 0
	for _, t := range w.Teams {
		objcount += len(t.GameObjs)
	}

	rtn.PartSize = int(math.Pow(float64(objcount), 1.0/3.0))
	if rtn.PartSize < 2 {
		rtn.PartSize = 2
	}
	//log.Printf("partsize:%v objs:%v", rtn.PartSize, objcount)

	rtn.refs = make([][][]GameObjectList, rtn.PartSize)
	for i := 0; i < rtn.PartSize; i++ {
		rtn.refs[i] = make([][]GameObjectList, rtn.PartSize)
		for j := 0; j < rtn.PartSize; j++ {
			rtn.refs[i][j] = make([]GameObjectList, rtn.PartSize)
		}
	}

	for _, t := range w.Teams {
		for _, obj := range t.GameObjs {
			if obj != nil && obj.objType != 0 {
				//if obj != nil {
				partPos := rtn.GetPartPos(&obj.posVector)
				rtn.AddPartPos(partPos, obj)
			}
		}
	}
	return &rtn
}
