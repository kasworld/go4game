package shootbase

import (
	"github.com/kasworld/go4game"
	//"encoding/json"
	//"errors"
	"fmt"
	//"log"
)

type GameObjectDisp struct {
	ID int64
	P  [3]int32
	R  int
	// P  go4game.Vector3D
	// R  float64
}

func NewGameObjectDisp(o *GameObject) *GameObjectDisp {
	gi := GameObjectDisp{
		ID: o.ID,
		P:  o.PosVector.NewInt32Vector(),
		R:  int(GameConst.Radius[o.ObjType]),
	}
	return &gi
}

type TeamDisp struct {
	ID     int64
	Color  int
	GOList []GameObjectDisp
}

func NewTeamDisp(t *Team) *TeamDisp {
	ts := TeamDisp{
		ID:     t.ID,
		Color:  t.Color,
		GOList: make([]GameObjectDisp, 0, len(t.GameObjs)),
	}
	for _, o := range t.GameObjs {
		if o.enabled {
			ts.GOList = append(ts.GOList, *NewGameObjectDisp(o))
		}

	}
	return &ts
}

type WorldDisp struct {
	ID       int64
	B1       *go4game.HyperRect
	B2       *go4game.HyperRect
	TeamList []TeamDisp
}

func NewWorldDisp(w *World) *WorldDisp {
	ws := WorldDisp{
		ID: w.ID,
		B1: GameConst.WorldCube,
		B2: GameConst.WorldCube2,

		TeamList: make([]TeamDisp, 0, len(w.Teams)),
	}
	for _, t := range w.Teams {
		ws.TeamList = append(ws.TeamList, *NewTeamDisp(t))
	}
	return &ws
}

// packet type
type PacketType int

const (
	_ PacketType = iota
	// for ai
	ReqNearInfo
	RspNearInfo
	// for observer
	ReqWorldInfo
	RspWorldInfo
)

type RspGamePacket struct {
	Cmd       PacketType
	TeamInfo  *TeamInfoPacket
	WorldInfo *WorldDisp
	NearObjs  SPObjList
}

func (gp RspGamePacket) String() string {
	return fmt.Sprintf("RspGamePacket Cmd:%v TeamInfo:%v WorldInfo:%v",
		gp.Cmd,
		gp.TeamInfo,
		gp.WorldInfo)
}

type ReqGamePacket struct {
	Cmd       PacketType
	ClientAct *ClientActionPacket
}

func (gp ReqGamePacket) String() string {
	return fmt.Sprintf("ReqGamePacket Cmd:%v ClientAct:%v",
		gp.Cmd,
		gp.ClientAct)
}

type TeamInfoPacket struct {
	*SPObj
	ActionPoint int
	Score       int
	HomePos     go4game.Vector3D
}

type ClientActionPacket struct {
	Accel           *go4game.Vector3D
	NormalBulletMv  *go4game.Vector3D
	BurstShot       int
	HommingTargetID go4game.IDList
	SuperBulletMv   *go4game.Vector3D
}
