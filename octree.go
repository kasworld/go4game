package go4game

import (
//"log"
//"math"
)

const (
	MaxOctreeData = 8
)

type Octree struct {
	BoundCube HyperRect
	Center    Vector3D
	DataList  SPObjList
	Children  [8]*Octree
}

func NewOctree(cube HyperRect) *Octree {
	rtn := Octree{
		BoundCube: cube,
		DataList:  make(SPObjList, 0, MaxOctreeData),
		Center:    cube.Center(),
	}
	//log.Printf("new octree %v", rtn.BoundCube)
	return &rtn
}

func MakeOctree(w *World) *Octree {
	//log.Printf("make octree")
	rtn := NewOctree(GameConst.WorldCube)
	for _, t := range w.Teams {
		for _, obj := range t.GameObjs {
			if obj != nil && obj.ObjType != 0 {
				rtn.Insert(NewSPObj(obj))
			}
		}
	}
	return rtn
}

func (ot *Octree) Split() {
	if ot.Children[0] != nil {
		return
	}
	// split all data and make datalist nil
	//log.Printf("split octree %v", ot.Center)
	for i, _ := range ot.Children {
		newbound := *ot.BoundCube.MakeCubeBy8Driect(ot.Center, i)
		ot.Children[i] = NewOctree(newbound)
	}
}

func (ot *Octree) Insert(o *SPObj) bool {
	//log.Printf("insert to octree obj%v %v", o.ID, o.PosVector)
	if !o.PosVector.IsIn(&ot.BoundCube) {
		//log.Printf("invalid Insert Octree %v %v", ot.BoundCube, o.PosVector)
		return false
	}
	if len(ot.DataList) < MaxOctreeData {
		// simple append
		ot.DataList = append(ot.DataList, o)
		return true
	} else {
		ot.Split()
		d8 := ot.Center.To8Direct(o.PosVector)
		return ot.Children[d8].Insert(o)
	}
}

func (ot *Octree) QueryByLen(fn CheckSPObjListFn, center Vector3D, r float64) bool {
	return ot.QueryByHyperRect(fn, NewHyperRectByCR(center, r))

	// if !ot.BoundCube.IsContact(center, r) {
	// 	return false
	// }
	// quit := fn(ot.DataList)
	// if quit {
	// 	return true
	// }
	// if ot.Children[0] == nil {
	// 	return false
	// }
	// for _, o := range ot.Children {
	// 	quit := o.QueryByLen(fn, center, r)
	// 	if quit {
	// 		return true
	// 	}
	// }
	// return false
}

func (ot *Octree) QueryByHyperRect(fn CheckSPObjListFn, hr *HyperRect) bool {
	if !ot.BoundCube.IsOverlap(hr) {
		return false
	}
	quit := fn(ot.DataList)
	if quit {
		return true
	}
	if ot.Children[0] == nil {
		return false
	}
	for _, o := range ot.Children {
		quit := o.QueryByHyperRect(fn, hr)
		if quit {
			return true
		}
	}
	return false
}
