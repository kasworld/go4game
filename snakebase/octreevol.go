package snakebase

import (
	//"log"
	//"math"
	//"time"
	"github.com/kasworld/go4game"
)

const (
	MaxOctreeVolData = 8
)

type OctreeVolObjI interface {
	Vol() *go4game.HyperRect
}

type OctreeVol struct {
	BoundCube *go4game.HyperRect
	Center    go4game.Vector3D
	DataList  []OctreeVolObjI
	Children  [8]*OctreeVol
}

func NewOctreeVol(cube *go4game.HyperRect) *OctreeVol {
	rtn := OctreeVol{
		BoundCube: cube,
		Center:    cube.Center(),
		DataList:  make([]OctreeVolObjI, 0, MaxOctreeVolData),
	}
	return &rtn
}

func (ot *OctreeVol) InsertChild(o OctreeVolObjI) bool {
	for _, chot := range ot.Children { // try child
		if chot.Insert(o) {
			return true
		}
	}
	return false
}

func (ot *OctreeVol) Insert(o OctreeVolObjI) bool {
	if !o.Vol().IsIn(ot.BoundCube) {
		return false
	}
	if ot.Children[0] != nil { // splited
		if !ot.InsertChild(o) { // append to me
			ot.DataList = append(ot.DataList, o)
		}
		return true
	} else { // not splited
		if len(ot.DataList) < MaxOctreeVolData { // check need split
			// simple append
			ot.DataList = append(ot.DataList, o)
			return true
		} else {
			ot.Split()
			if !ot.InsertChild(o) { // append to me
				ot.DataList = append(ot.DataList, o)
			}
			return true
		}
	}
}

func (ot *OctreeVol) Split() {
	if ot.Children[0] != nil {
		return
	}
	// split all data and make datalist nil
	//log.Printf("split octree %v %v", ot.BoundCube, ot.Center)
	for i, _ := range ot.Children {
		newbound := ot.BoundCube.MakeCubeBy8Driect(ot.Center, i)
		ot.Children[i] = NewOctreeVol(newbound)
	}
	// move this node data to child
	newDataList := make([]OctreeVolObjI, 0, len(ot.DataList))
	for _, o := range ot.DataList {
		if !ot.InsertChild(o) {
			newDataList = append(newDataList, o)
		}
	}
	ot.DataList = newDataList
}

func (ot *OctreeVol) QueryByHyperRect(fn func(OctreeVolObjI) bool, hr *go4game.HyperRect) bool {
	if !ot.BoundCube.IsOverlap(hr) {
		return false
	}
	for _, o := range ot.DataList {
		if !o.Vol().IsOverlap(hr) {
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
