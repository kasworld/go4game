package snakebase

import (
	//"encoding/json"
	"fmt"
	"github.com/kasworld/go4game"
	//"log"
	//"os"
	//"runtime"
	"math/rand"
	"time"
)

type InteractType int

const (
	NoInteract InteractType = iota
	InteractRemove
	InteractBlock
)

type GameObjBase struct {
	id      int64
	GroupID int64
	Color   uint32
}

func MakeGameObjBase(og ObjGroupI, color uint32) GameObjBase {
	o := GameObjBase{
		id:      <-go4game.IdGenCh,
		GroupID: og.ID(),
		Color:   color,
	}
	return o
}

type SnakeHead struct {
	GameObjBase
	MoveVector go4game.Vector3D
	Sphere
}

func NewSnakeHead(og ObjGroupI) *SnakeHead {
	o := SnakeHead{
		GameObjBase: MakeGameObjBase(og, rand.Uint32()),
		MoveVector:  SnakeDefault.WorldCube.RandVector().NormalizedTo(20),
	}
	return &o
}

func (o *SnakeHead) ID() int64 {
	return o.id
}
func (o SnakeHead) String() string {
	return fmt.Sprintf("SnakeHead ID:%v Group%v", o.ID, o.GroupID)
}
func (o *SnakeHead) Vol() *go4game.HyperRect {
	return o.Sphere.Vol()
}
func (o *SnakeHead) ToOctreeVolObj() OctreeVolObjI {
	rtn := *o
	return &rtn
}
func (o *SnakeHead) ActByTime(w WorldI, t time.Time) {
	o.MoveVector.NormalizedTo(20)
	o.Sphere.center = o.Sphere.center.Add(o.MoveVector)
}

type SnakeTail struct {
	GameObjBase
	Sphere
	enabled bool
	endTime time.Time
}

func NewSnakeTail(og ObjGroupI, pos go4game.Vector3D) *SnakeTail {
	o := SnakeTail{
		GameObjBase: MakeGameObjBase(og, pos),
		endTime:     time.Now().Add(time.Second * 10),
	}
	return &o
}

func (o *SnakeTail) ID() int64 {
	return o.id
}
func (o SnakeTail) String() string {
	return fmt.Sprintf("SnakeTail ID:%v Group%v", o.ID, o.GroupID)
}
func (o *SnakeTail) Vol() *go4game.HyperRect {
	return o.Sphere.Vol()
}
func (o *SnakeTail) ToOctreeVolObj() OctreeVolObjI {
	rtn := *o
	return &rtn
}
func (o *SnakeTail) ActByTime(w WorldI, t time.Time) {
	if o.endTime.Before(t) {
		o.enabled = false
	}
}

type Plum struct {
	GameObjBase
	MoveVector go4game.Vector3D
	Sphere
}

func NewPlum(og ObjGroupI) *Plum {
	o := Plum{
		GameObjBase: MakeGameObjBase(og, rand.Uint32()),
		MoveVector:  SnakeDefault.WorldCube.RandVector().NormalizedTo(20),
	}
	return &o
}
func (o *Plum) ID() int64 {
	return o.id
}
func (o Plum) String() string {
	return fmt.Sprintf("Plum ID:%v Group%v", o.ID, o.GroupID)
}
func (o *Plum) Vol() *go4game.HyperRect {
	return o.Sphere.Vol()
}
func (o *Plum) ToOctreeVolObj() OctreeVolObjI {
	rtn := *o
	return &rtn
}
func (o *Plum) ActByTime(w WorldI, t time.Time) {
	o.MoveVector.NormalizedTo(20)
	o.Sphere.center = o.Sphere.center.Add(o.MoveVector)
}

type Apple struct {
	GameObjBase
	Sphere
}

func NewApple(og ObjGroupI) *Apple {
	o := Apple{
		GameObjBase: MakeGameObjBase(og, rand.Uint32()),
	}
	return &o
}
func (o *Apple) ID() int64 {
	return o.id
}
func (o Apple) String() string {
	return fmt.Sprintf("Apple ID:%v Group%v", o.ID, o.GroupID)
}
func (o *Apple) Vol() *go4game.HyperRect {
	return o.Sphere.Vol()
}
func (o *Apple) ToOctreeVolObj() OctreeVolObjI {
	rtn := *o
	return &rtn
}
func (o *Apple) ActByTime(w WorldI, t time.Time) {
}

type Wall struct {
	GameObjBase
	Cube
}

func NewWall(og ObjGroupI) *Wall {
	o := Wall{
		GameObjBase: MakeGameObjBase(og, rand.Uint32()),
	}
	return &o
}
func (o *Wall) ID() int64 {
	return o.id
}
func (o Wall) String() string {
	return fmt.Sprintf("Wall ID:%v Group%v", o.ID, o.GroupID)
}
func (o *Wall) Vol() *go4game.HyperRect {
	return o.Cube.Vol()
}
func (o *Wall) ToOctreeVolObj() OctreeVolObjI {
	rtn := *o
	return &rtn
}
func (o *Wall) ActByTime(w WorldI, t time.Time) {
}

func test_GameObjI() {
	var o GameObjI
	//o = &GameObjBase{}
	o = &SnakeHead{}
	o = &SnakeTail{}
	o = &Plum{}
	o = &Apple{}
	o = &Wall{}
	_ = o
}
