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

// type GameConfigI interface {
// 	Validate()
// 	Save(filename string) bool
// 	Load(filename string) bool
// 	SaveLoad(filename string)
// }

// type ServiceI interface {
// 	CmdReceiver
// 	ID() int64
// 	AddWorld(WorldI)
// 	RemoveWorld(id int64)
// }

type WorldI interface {
	CmdReceiver
	ID() int64
	AddObjGroup(ObjGroupI)
	RemoveObjGroup(id int64)
	CollideList(o GameObjI) []GameObjI
}

type ObjGroupI interface {
	ID() int64
	AddGameObj(GameObjI)
	RemoveGameObj(id int64)
	StartFrameAction(w WorldI, ftime time.Time)
	FrameActionResult() interface{}
	AddInitMembers()
}

type OGActor interface {
	FrameAction() go4game.Vector3D
}

type GameObjI interface {
	ID() int64
	OctreeVolObjI
	ToOctreeVolObj() OctreeVolObjI
	ActByTime(w WorldI, t time.Time)
}

func test_interfaces() {
	test_WorldI()
	test_ObjGroupI()
	test_GameObjI()
	test_Collider()
}

func init() {
	test_interfaces()
}
