package snakebase

import (
	//"encoding/json"
	"fmt"
	"github.com/kasworld/go4game"
	//"log"
	//"os"
	//"runtime"
	"time"
)

type GameObjBase struct {
	id           int64
	GroupID      int64
	PosVector    go4game.Vector3D
	InteractType int
}

func (o *GameObjBase) ID() int64 {
	return o.id
}
func (o GameObjBase) String() string {
	return fmt.Sprintf("ID:%v Group%v", o.ID, o.GroupID)
}
func (o *GameObjBase) Pos() go4game.Vector3D {
	return o.PosVector
}
func (o *GameObjBase) ToOctreeObj() go4game.OctreeObjI {
	if o == nil {
		return nil
	}
	rtn := *o
	return &rtn
}
func (o *GameObjBase) ActByTime(w WorldI, t time.Time) {
}

type SnakeHead struct {
	GameObjBase
	MoveVector go4game.Vector3D
}

func NewSnakeHead(og ObjGroupI) *SnakeHead {
	o := SnakeHead{
		GameObjBase: GameObjBase{
			id:           <-go4game.IdGenCh,
			GroupID:      og.ID(),
			PosVector:    GameConst.WorldCube.RandVector(),
			InteractType: 1,
		},
		MoveVector: GameConst.WorldCube.RandVector().NormalizedTo(20),
	}
	return &o
}

func (o *SnakeHead) ID() int64 {
	return o.id
}
func (o SnakeHead) String() string {
	return fmt.Sprintf("SnakeHead ID:%v Group%v", o.ID, o.GroupID)
}
func (o *SnakeHead) Pos() go4game.Vector3D {
	return o.PosVector
}
func (o *SnakeHead) ToOctreeObj() go4game.OctreeObjI {
	return o.GameObjBase.ToOctreeObj()
}
func (o *SnakeHead) ActByTime(w WorldI, t time.Time) {
	o.MoveVector.NormalizedTo(20)
	o.PosVector = o.PosVector.Add(o.MoveVector)
}

type SnakeTail struct {
	GameObjBase
	enabled bool
	endTime time.Time
}

func NewSnakeTail(og ObjGroupI, pos go4game.Vector3D) *SnakeTail {
	o := SnakeTail{
		GameObjBase: GameObjBase{
			id:           <-go4game.IdGenCh,
			GroupID:      og.ID(),
			PosVector:    pos,
			InteractType: 1,
		},
		endTime: time.Now().Add(time.Second * 10),
	}
	return &o
}

func (o *SnakeTail) ID() int64 {
	return o.id
}
func (o SnakeTail) String() string {
	return fmt.Sprintf("SnakeTail ID:%v Group%v", o.ID, o.GroupID)
}
func (o *SnakeTail) Pos() go4game.Vector3D {
	return o.PosVector
}
func (o *SnakeTail) ToOctreeObj() go4game.OctreeObjI {
	return o.GameObjBase.ToOctreeObj()
}
func (o *SnakeTail) ActByTime(w WorldI, t time.Time) {
	if o.endTime.Before(t) {
		o.enabled = false
	}
}

type Plum struct {
	GameObjBase
	MoveVector go4game.Vector3D
}

func NewPlum(og ObjGroupI) *Plum {
	o := Plum{
		GameObjBase: GameObjBase{
			id:           <-go4game.IdGenCh,
			GroupID:      og.ID(),
			PosVector:    GameConst.WorldCube.RandVector(),
			InteractType: 1,
		},
		MoveVector: GameConst.WorldCube.RandVector().NormalizedTo(20),
	}
	return &o
}
func (o *Plum) ID() int64 {
	return o.id
}
func (o Plum) String() string {
	return fmt.Sprintf("Plum ID:%v Group%v", o.ID, o.GroupID)
}
func (o *Plum) Pos() go4game.Vector3D {
	return o.PosVector
}
func (o *Plum) ToOctreeObj() go4game.OctreeObjI {
	return o.GameObjBase.ToOctreeObj()
}
func (o *Plum) ActByTime(w WorldI, t time.Time) {
	o.MoveVector.NormalizedTo(20)
	o.PosVector = o.PosVector.Add(o.MoveVector)
}

type Apple struct {
	GameObjBase
}

func NewApple(og ObjGroupI) *Apple {
	o := Apple{
		GameObjBase: GameObjBase{
			id:           <-go4game.IdGenCh,
			GroupID:      og.ID(),
			PosVector:    GameConst.WorldCube.RandVector(),
			InteractType: 1,
		},
	}
	return &o
}
func (o *Apple) ID() int64 {
	return o.id
}
func (o Apple) String() string {
	return fmt.Sprintf("Apple ID:%v Group%v", o.ID, o.GroupID)
}
func (o *Apple) Pos() go4game.Vector3D {
	return o.PosVector
}
func (o *Apple) ToOctreeObj() go4game.OctreeObjI {
	return o.GameObjBase.ToOctreeObj()
}
func (o *Apple) ActByTime(w WorldI, t time.Time) {
}

type Wall struct {
	GameObjBase
}

func NewWall(og ObjGroupI) *Wall {
	o := Wall{
		GameObjBase: GameObjBase{
			id:           <-go4game.IdGenCh,
			GroupID:      og.ID(),
			PosVector:    GameConst.WorldCube.RandVector(),
			InteractType: 1,
		},
	}
	return &o
}
func (o *Wall) ID() int64 {
	return o.id
}
func (o Wall) String() string {
	return fmt.Sprintf("Wall ID:%v Group%v", o.ID, o.GroupID)
}
func (o *Wall) Pos() go4game.Vector3D {
	return o.PosVector
}
func (o *Wall) ToOctreeObj() go4game.OctreeObjI {
	return o.GameObjBase.ToOctreeObj()
}
func (o *Wall) ActByTime(w WorldI, t time.Time) {
}
