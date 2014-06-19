package shootbase

import (
	"fmt"
	"github.com/kasworld/go4game"
)

type GameObjectType int

type SPObj struct {
	ID         int64
	TeamID     int64
	PosVector  go4game.Vector3D
	MoveVector go4game.Vector3D
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

func (o SPObj) Pos() go4game.Vector3D {
	return o.PosVector
}
