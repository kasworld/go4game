package snakebase

import (
	// "fmt"
	"github.com/kasworld/go4game"
	// "runtime"
	"time"
)

type CmdReceiver interface {
	SendGoCmd(Cmd string, Args interface{}, Rsp chan<- interface{})
	Loop()
}

type GameConfigI interface {
	Validate()
	Save(filename string) bool
	Load(filename string) bool
	SaveLoad(filename string)
	NewService() ServiceI
}

type ServiceI interface {
	CmdReceiver
	ID() int64
	NewWorld() WorldI
	AddWorld(WorldI)
	RemoveWorld(id int64)
}

type WorldI interface {
	CmdReceiver
	ID() int64
	AddObjGroup(ObjGroupI)
	RemoveObjGroup(id int64)
}

type ObjGroupI interface {
	ID() int64
	AddGameObj(GameObjI)
	RemoveGameObj(id int64)
	DoFrameAction(ftime time.Time) <-chan interface{}
	AddInitMembers()
}

type OGActor interface {
	FrameAction() go4game.Vector3D
}

type GameObjI interface {
	ID() int64
	go4game.OctreeObjI
	ToOctreeObj() go4game.OctreeObjI
	ActByTime(w WorldI, t time.Time)
}
