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
	String() string
	NewWorld() WorldI
}

type WorldI interface {
	CmdReceiver
	String() string
	AddObjGroup(ObjGroupI)
	RemoveObjGroup(id int64)
}

type ObjGroupI interface {
	AddGameObj(GameObjI)
	RemoveGameObj(id int64)
	DoFrameAction(ftime time.Time) <-chan interface{}
}

type GameObjI interface {
	go4game.OctreeObjI
	ToOctreeObj() go4game.OctreeObjI
	ActByTime(WorldI, t time.Time) go4game.IDList
	// MoveByTime(envInfo *ActionFnEnvInfo) bool
	// BorderAction(envInfo *ActionFnEnvInfo) bool
	// CollisionAction(envInfo *ActionFnEnvInfo) bool
	// ExpireAction(envInfo *ActionFnEnvInfo) bool
}
