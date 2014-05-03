package go4game

import (
	//"encoding/json"
	//"errors"
	"fmt"
	//"log"
)

type GameObjectSeiralize struct {
	ID              int
	ObjType         GameObjectType
	PosVector       Vector3D
	MoveVector      Vector3D
	CollisionRadius float64
}

func NewGameObjectSeiralize(o *GameObject) *GameObjectSeiralize {
	gi := GameObjectSeiralize{
		ID:              o.ID,
		ObjType:         o.ObjType,
		PosVector:       o.PosVector,
		MoveVector:      o.MoveVector,
		CollisionRadius: o.CollisionRadius,
	}
	//log.Printf("%#v", gi)
	return &gi
}

type TeamSeialize struct {
	ID     int
	Color  int
	GOList []GameObjectSeiralize
}

func NewTeamSeialize(t *Team) *TeamSeialize {
	ts := TeamSeialize{
		ID:     t.ID,
		Color:  t.Color,
		GOList: make([]GameObjectSeiralize, 0, len(t.GameObjs)),
	}
	for _, o := range t.GameObjs {
		if o.enabled {
			ts.GOList = append(ts.GOList, *NewGameObjectSeiralize(o))
		}

	}
	return &ts
}

type WorldSerialize struct {
	ID       int
	MinPos   Vector3D
	MaxPos   Vector3D
	TeamList []TeamSeialize
}

func NewWorldSerialize(w *World) *WorldSerialize {
	ws := WorldSerialize{
		ID:       w.ID,
		MinPos:   w.MinPos,
		MaxPos:   w.MaxPos,
		TeamList: make([]TeamSeialize, 0, len(w.Teams)),
	}
	for _, t := range w.Teams {
		ws.TeamList = append(ws.TeamList, *NewTeamSeialize(t))
	}
	return &ws
}

const (
	_ = iota
	// ReqMakeTeam
	// RspMakeTeam
	ReqFrameInfo
	RspFrameInfo
	ReqWorldInfo
	RspWorldInfo
	// ReqSpatialPartition
	// RspSpatialPartition
	ReqAIAct
	RspAIAct
)

type GamePacket struct {
	Cmd       int
	TeamInfo  *TeamInfoPacket
	WorldInfo *WorldSerialize
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
	SPObj
}

type ClientActionPacket struct {
	Accel           *Vector3D
	NormalBulletMv  *Vector3D
	HommingTargetID int
	SuperBulletMv   *Vector3D
}
