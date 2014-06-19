package go4game

import (
	"log"
	//"math"
	//"time"
)

const (
	MaxOctreeData = 8
)

type OctreeObj interface {
	Pos() Vector3D
	String() string
}

type OctreeObjList []OctreeObj

type Octree struct {
	BoundCube *HyperRect
	Center    Vector3D
	DataList  OctreeObjList
	Children  [8]*Octree
}

func NewOctree(cube *HyperRect) *Octree {
	rtn := Octree{
		BoundCube: cube,
		DataList:  make(OctreeObjList, 0, MaxOctreeData),
		Center:    cube.Center(),
	}
	//log.Printf("new octree %v", rtn.BoundCube)
	return &rtn
}

func (ot *Octree) Split() {
	if ot.Children[0] != nil {
		return
	}
	// split all data and make datalist nil
	//log.Printf("split octree %v %v", ot.BoundCube, ot.Center)
	for i, _ := range ot.Children {
		newbound := ot.BoundCube.MakeCubeBy8Driect(ot.Center, i)
		ot.Children[i] = NewOctree(newbound)
	}
}

func (ot *Octree) Insert(o OctreeObj) bool {
	//log.Printf("insert to octree obj%v %v", o.ID, o.Pos())
	if !o.Pos().IsIn(ot.BoundCube) {
		log.Printf("invalid Insert Octree %v %v", ot.BoundCube, o.Pos())
		return false
	}
	if len(ot.DataList) < MaxOctreeData {
		// simple append
		ot.DataList = append(ot.DataList, o)
		return true
	} else {
		ot.Split()
		d8 := ot.Center.To8Direct(o.Pos())
		return ot.Children[d8].Insert(o)
	}
}

func (ot *Octree) QueryByHyperRect(fn func(OctreeObj) bool, hr *HyperRect) bool {
	if !ot.BoundCube.IsOverlap(hr) {
		return false
	}
	for _, o := range ot.DataList {
		if !o.Pos().IsIn(hr) {
			continue
		}
		if fn(o) {
			return true
		}
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
