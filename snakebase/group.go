package snakebase

import (
	//"encoding/json"
	//"fmt"
	"github.com/kasworld/go4game"
	//"log"
	//"os"
	//"runtime"
	"math/rand"
	"time"
)

type ObjGroupBase struct {
	id       int64
	gameObjs map[int64]GameObjI
	chStep   chan interface{}
}

func NewObjGroup(w *World) *ObjGroupBase {
	og := ObjGroupBase{
		id:       <-go4game.IdGenCh,
		gameObjs: make(map[int64]GameObjI),
		chStep:   make(chan interface{}, 1),
	}
	return &og
}
func (og *ObjGroupBase) ID() int64 {
	return og.id
}
func (og *ObjGroupBase) AddGameObj(o GameObjI) {
	og.gameObjs[o.ID()] = o
}
func (og *ObjGroupBase) RemoveGameObj(id int64) {
	delete(og.gameObjs, id)
}
func (og *ObjGroupBase) StartFrameAction(world WorldI, ftime time.Time) {
	for _, o := range og.gameObjs {
		o.ActByTime(world, ftime)
	}
	og.chStep <- nil
}
func (og *ObjGroupBase) FrameActionResult() interface{} {
	return <-og.chStep
}
func (og *ObjGroupBase) AddInitMembers() {
}
func (og *ObjGroupBase) GameObjI() map[int64]GameObjI {
	return og.gameObjs
}
func (og *ObjGroupBase) AddToOctreeVol(ot *OctreeVol) {
	for _, o := range og.gameObjs {
		ot.Insert(o.ToOctreeVolObj())
	}
}

type Snake struct {
	ObjGroupBase
	Color  uint32
	HeadID int64
}

func NewSnake(w *World) *Snake {
	og := Snake{
		ObjGroupBase: *NewObjGroup(w),
		Color:        rand.Uint32(),
	}
	og.AddInitMembers()
	//log.Printf("%#v", og)
	return &og
}
func (og *Snake) ID() int64 {
	return og.id
}
func (og *Snake) AddGameObj(o GameObjI) {
	og.ObjGroupBase.AddGameObj(o)
}
func (og *Snake) RemoveGameObj(id int64) {
	og.ObjGroupBase.RemoveGameObj(id)
}
func (og *Snake) StartFrameAction(world WorldI, ftime time.Time) {
	for _, o := range og.gameObjs {
		o.ActByTime(world, ftime)
	}
	og.chStep <- nil
}
func (og *Snake) FrameActionResult() interface{} {
	return <-og.chStep
}
func (og *Snake) AddInitMembers() {
	o := NewSnakeHead(og)
	og.HeadID = o.ID()
	og.AddGameObj(o)
}
func (og *Snake) GameObjI() map[int64]GameObjI {
	return og.gameObjs
}
func (og *Snake) AddToOctreeVol(ot *OctreeVol) {
	og.ObjGroupBase.AddToOctreeVol(ot)
}

type StageWalls struct {
	ObjGroupBase
	ExitPos go4game.Vector3D
	Color   int
}

func NewStageWalls(w *World) *StageWalls {
	og := StageWalls{
		ObjGroupBase: *NewObjGroup(w),
	}
	og.AddInitMembers()
	return &og
}
func (og *StageWalls) ID() int64 {
	return og.id
}
func (og *StageWalls) AddGameObj(o GameObjI) {
	og.ObjGroupBase.AddGameObj(o)
}
func (og *StageWalls) RemoveGameObj(id int64) {
	og.ObjGroupBase.RemoveGameObj(id)
}
func (og *StageWalls) StartFrameAction(world WorldI, ftime time.Time) {
	for _, o := range og.gameObjs {
		o.ActByTime(world, ftime)
	}
	og.chStep <- nil
}
func (og *StageWalls) FrameActionResult() interface{} {
	return <-og.chStep
}
func (og *StageWalls) AddInitMembers() {
}
func (og *StageWalls) GameObjI() map[int64]GameObjI {
	return og.gameObjs
}
func (og *StageWalls) AddToOctreeVol(ot *OctreeVol) {
	og.ObjGroupBase.AddToOctreeVol(ot)
}

type StagePlums struct {
	ObjGroupBase
	Color int
}

func NewStagePlums(w *World) *StagePlums {
	og := StagePlums{
		ObjGroupBase: *NewObjGroup(w),
	}
	og.AddInitMembers()
	return &og
}
func (og *StagePlums) ID() int64 {
	return og.id
}
func (og *StagePlums) AddGameObj(o GameObjI) {
	og.ObjGroupBase.AddGameObj(o)
}
func (og *StagePlums) RemoveGameObj(id int64) {
	og.ObjGroupBase.RemoveGameObj(id)
}
func (og *StagePlums) StartFrameAction(world WorldI, ftime time.Time) {
	for _, o := range og.gameObjs {
		o.ActByTime(world, ftime)
	}
	og.chStep <- nil
}
func (og *StagePlums) FrameActionResult() interface{} {
	return <-og.chStep
}
func (og *StagePlums) AddInitMembers() {
}
func (og *StagePlums) GameObjI() map[int64]GameObjI {
	return og.gameObjs
}
func (og *StagePlums) AddToOctreeVol(ot *OctreeVol) {
	og.ObjGroupBase.AddToOctreeVol(ot)
}

type StageApples struct {
	ObjGroupBase
	Color int
}

func NewStageApples(w *World) *StageApples {
	og := StageApples{
		ObjGroupBase: *NewObjGroup(w),
	}
	og.AddInitMembers()
	return &og
}
func (og *StageApples) ID() int64 {
	return og.id
}
func (og *StageApples) AddGameObj(o GameObjI) {
	og.ObjGroupBase.AddGameObj(o)
}
func (og *StageApples) RemoveGameObj(id int64) {
	og.ObjGroupBase.RemoveGameObj(id)
}
func (og *StageApples) StartFrameAction(world WorldI, ftime time.Time) {
	for _, o := range og.gameObjs {
		o.ActByTime(world, ftime)
	}
	og.chStep <- nil
}
func (og *StageApples) FrameActionResult() interface{} {
	return <-og.chStep
}
func (og *StageApples) AddInitMembers() {
}
func (og *StageApples) GameObjI() map[int64]GameObjI {
	return og.gameObjs
}
func (og *StageApples) AddToOctreeVol(ot *OctreeVol) {
	og.ObjGroupBase.AddToOctreeVol(ot)
}

func test_ObjGroupI() {
	var og ObjGroupI
	og = &ObjGroupBase{}
	og = &Snake{}
	og = &StageWalls{}
	og = &StagePlums{}
	og = &StageApples{}
	_ = og
}
