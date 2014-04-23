package go4game

import (
	//"encoding/json"
	//"errors"
	"fmt"
)

type GameObjectSeiralize struct {
	ID              int
	objType         string
	posVector       Vector3D
	moveVector      Vector3D
	collisionRadius float64
}

func NewGameObjectSeiralize(o *GameObject) *GameObjectSeiralize {
	gi := GameObjectSeiralize{
		ID:              o.ID,
		objType:         o.objType,
		posVector:       o.posVector,
		moveVector:      o.moveVector,
		collisionRadius: o.collisionRadius,
	}
	return &gi
}

type TeamSeialize struct {
	ID     int
	GOList []GameObjectSeiralize
}

func NewTeamSeialize(t *Team) *TeamSeialize {
	ts := TeamSeialize{
		ID:     t.ID,
		GOList: make([]GameObjectSeiralize, len(t.GameObjs)),
	}
	for _, o := range t.GameObjs {
		ts.GOList = append(ts.GOList, *NewGameObjectSeiralize(o))
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
		TeamList: make([]TeamSeialize, len(w.Teams)),
	}
	for _, t := range w.Teams {
		ws.TeamList = append(ws.TeamList, *NewTeamSeialize(t))
	}
	return &ws
}

const (
	_ = iota
	ReqMakeTeam
	RspMakeTeam
	ReqWorldInfo
	RspWorldInfo
	ReqAIAct
	RspAIAct
)

type GamePacket struct {
	Cmd       int
	TeamInfo  *TeamInfoPacket
	WorldInfo *WorldSerialize
	AiAct     *AiActionPacket
}

func (gp GamePacket) String() string {
	return fmt.Sprintf("GamePacket Cmd:%v TeamInfo:%v WorldInfo:%v AiAct:%v",
		gp.Cmd,
		gp.TeamInfo,
		gp.WorldInfo,
		gp.AiAct)
}

type TeamInfoPacket struct {
	Teamname      string
	Teamcolor     []int
	Teamid        int
	TeamStartTime int
}

type AiActionPacket struct {
	Accel         Vector3D
	Fire1TargetID int
	Fire2TargetID int
	Fire3TargetID int
}

// func parsePacket(buf []byte) (interface{}, error) {
// 	var pk CmdPacket
// 	err := json.Unmarshal(buf, &pk)
// 	if err != nil {
// 		return nil, err
// 	}
// 	switch pk.Cmd {
// 	case "makeTeam":
// 		var pk MakeTeamPacket
// 		err := json.Unmarshal(buf, &pk)
// 		return &pk, err
// 	case "aiAction":
// 		var pk AiActionPacket
// 		err := json.Unmarshal(buf, &pk)
// 		return &pk, err
// 	case "worldInfo":
// 		var pk WorldInfoPacket
// 		err := json.Unmarshal(buf, &pk)
// 		return &pk, err
// 	default:
// 		errmsg := fmt.Sprintf("unknown packet %v", pk.Cmd)
// 		return nil, errors.New(errmsg)
// 	}
// }
