package go4game

import (
	"math"
)

type SpatialPartition struct {
	Min, Max Vector3D
	PartSize int
	refs     [][][]GameObjectList
}

func (p *SpatialPartition) GetPartPos(pos *Vector3D) [3]int {
	rtn := [3]int{0, 0, 0}

	for i, v := range pos {
		l := p.Max[i] - p.Min[i]
		rtn[i] = int((v - p.Min[i]) / l * float64(p.PartSize))
		if rtn[i] >= p.PartSize {
			rtn[i] = p.PartSize - 1
		}
		if rtn[i] < 0 {
			rtn[i] = 0
		}
	}
	return rtn
}

func (p *SpatialPartition) AddPartPos(pos [3]int, obj *GameObject) {
	p.refs[pos[0]][pos[1]][pos[2]] = append(p.refs[pos[0]][pos[1]][pos[2]], obj)
}

func (p *SpatialPartition) GetNear(pos *Vector3D) GameObjectList {
	ppos := p.GetPartPos(pos)
	return p.refs[ppos[0]][ppos[1]][ppos[2]]
}

func (p *SpatialPartition) GetNear2(pos *Vector3D) GameObjectList {
	ppos := p.GetPartPos(pos)
	rtngl := make(GameObjectList, 0)
	for i := ppos[0] - 1; i < ppos[0]+1; i++ {
		if i < 0 || i >= p.PartSize {
			continue
		}
		for j := ppos[0] - 1; j < ppos[0]+1; j++ {
			if j < 0 || j >= p.PartSize {
				continue
			}
			for k := ppos[0] - 1; k < ppos[0]+1; k++ {
				if k < 0 || k >= p.PartSize {
					continue
				}
				rtngl = append(rtngl, p.refs[i][j][k]...)
			}
		}
	}
	return rtngl
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

	rtn.refs = make([][][]GameObjectList, rtn.PartSize)
	for i := 0; i < rtn.PartSize; i++ {
		rtn.refs[i] = make([][]GameObjectList, rtn.PartSize)
		for j := 0; j < rtn.PartSize; j++ {
			rtn.refs[i][j] = make([]GameObjectList, rtn.PartSize)
		}
	}

	for _, t := range w.Teams {
		for _, obj := range t.GameObjs {
			if obj != nil && obj.enabled {
				partPos := rtn.GetPartPos(&obj.posVector)
				rtn.AddPartPos(partPos, obj)
			}
		}
	}
	return &rtn
}
