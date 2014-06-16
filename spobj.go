package go4game

import (
	"fmt"
)

type GameObjectType int

type SPObj struct {
	ID         int64
	TeamID     int64
	PosVector  Vector3D
	MoveVector Vector3D
	ObjType    GameObjectType
}

func (o *SPObj) IsCollision(s *SPObj) bool {
	if GameConst.IsInteract[o.ObjType][s.ObjType] &&
		(s.PosVector.Sqd(o.PosVector) <= GameConst.ObjSqd[s.ObjType][o.ObjType]) {
		return true
	}
	return false
}

type SPObjList []*SPObj

func (o SPObj) String() string {
	return fmt.Sprintf("ID:%v Type:%v Team%v", o.ID, o.ObjType, o.TeamID)
}
