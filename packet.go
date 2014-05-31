package go4game

import (
	//"encoding/json"
	//"errors"
	"fmt"
	//"log"
)

type GameObjectDisp struct {
	ID int64
	P  [3]int32
	R  int
	// P  Vector3D
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
	TeamList []TeamDisp
}

func NewWorldDisp(w *World) *WorldDisp {
	ws := WorldDisp{
		ID:       w.ID,
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
	ReqFrameInfo
	RspFrameInfo
	// for observer
	ReqWorldInfo
	RspWorldInfo
	// ReqAIAct
	// RspAIAct
)

type GamePacket struct {
	Cmd       PacketType
	TeamInfo  *TeamInfoPacket
	WorldInfo *WorldDisp
	ClientAct *ClientActionPacket
	Spp       *SpatialPartition
}

func (gp GamePacket) String() string {
	return fmt.Sprintf("GamePacket Cmd:%v TeamInfo:%v WorldInfo:%v ClientAct:%v",
		gp.Cmd,
		gp.TeamInfo,
		gp.WorldInfo,
		gp.ClientAct)
}

type TeamInfoPacket struct {
	*SPObj
	ActionPoint int
	Score       int
	HomePos     Vector3D
}

type ClientActionPacket struct {
	Accel           *Vector3D
	NormalBulletMv  *Vector3D
	BurstShot       int
	HommingTargetID IDList
	SuperBulletMv   *Vector3D
}
