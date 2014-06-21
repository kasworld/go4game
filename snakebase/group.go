package snakebase

import (
	//"encoding/json"
	//"fmt"
	"github.com/kasworld/go4game"
	//"log"
	//"os"
	//"runtime"
	"time"
)

type ObjGroupBase struct {
	id       int64
	GameObjs map[int64]GameObjI
	config   *SnakeConfig
}

func (og *ObjGroupBase) ID() int64 {
	return og.id
}
func (og *ObjGroupBase) AddGameObj(o GameObjI) {
	og.GameObjs[o.ID()] = o
}
func (og *ObjGroupBase) RemoveGameObj(id int64) {
	delete(og.GameObjs, id)
}
func (og *ObjGroupBase) DoFrameAction(ftime time.Time) <-chan interface{} {
	rtn := make(chan interface{}, 1)
	return rtn
}
func (og *ObjGroupBase) AddInitMembers() {
}

type Snake struct {
	ObjGroupBase
	OGActor
	Color  int
	HeadID int64
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
func (og *Snake) DoFrameAction(ftime time.Time) <-chan interface{} {
	rtn := make(chan interface{}, 1)
	return rtn
}
func (og *Snake) AddInitMembers() {
	o := SnakeHead{
		GameObjBase: GameObjBase{
			id:           <-go4game.IdGenCh,
			GroupID:      og.ID(),
			PosVector:    og.config.WorldCube.RandVector(),
			InteractType: 1,
		},
		MoveVector: og.config.WorldCube.RandVector().NormalizedTo(20),
	}
	og.AddGameObj(&o)
}

type StageWalls struct {
	ObjGroupBase
	ExitPos go4game.Vector3D
	Color   int
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
func (og *StageWalls) DoFrameAction(ftime time.Time) <-chan interface{} {
	rtn := make(chan interface{}, 1)
	return rtn
}
func (og *StageWalls) AddInitMembers() {
}

type StagePlums struct {
	ObjGroupBase
	Color int
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
func (og *StagePlums) DoFrameAction(ftime time.Time) <-chan interface{} {
	rtn := make(chan interface{}, 1)
	return rtn
}
func (og *StagePlums) AddInitMembers() {
}

type StageApples struct {
	ObjGroupBase
	Color int
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
func (og *StageApples) DoFrameAction(ftime time.Time) <-chan interface{} {
	rtn := make(chan interface{}, 1)
	return rtn
}
func (og *StageApples) AddInitMembers() {
}
