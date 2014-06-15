package go4game

import (
	"time"
)

const (
	GameObjNil GameObjectType = iota
	GameObjMain
	GameObjShield
	GameObjBullet
	GameObjHommingBullet
	GameObjSuperBullet
	GameObjDeco
	GameObjMark
	GameObjHard
	GameObjFood
	GameObjEnd
)

func (o *GameObject) MakeMainObj() *GameObject {
	o.PosVector = GameConst.WorldCube.RandVector()
	o.MoveVector = GameConst.WorldCube.RandVector()
	o.accelVector = GameConst.WorldCube.RandVector()
	//o.borderActionFn = borderActionFn_Block
	o.borderActionFn = borderActionFn_Bounce
	o.ObjType = GameObjMain
	return o
}
func (o *GameObject) MakeShield(mo *GameObject) *GameObject {
	o.MoveVector = GameConst.WorldCube.RandVector()
	o.accelVector = GameConst.WorldCube.RandVector()
	o.moveByTimeFn = moveByTimeFn_shield
	o.borderActionFn = borderActionFn_B2_Disable
	o.PosVector = mo.PosVector
	o.ObjType = GameObjShield
	return o
}
func (o *GameObject) MakeBullet(mo *GameObject, MoveVector Vector3D) *GameObject {
	o.endTime = o.startTime.Add(time.Second * 10)
	o.PosVector = mo.PosVector
	o.MoveVector = MoveVector
	o.borderActionFn = borderActionFn_B2_Disable
	o.accelVector = Vector3D{0, 0, 0}
	o.ObjType = GameObjBullet
	return o
}
func (o *GameObject) MakeSuperBullet(mo *GameObject, MoveVector Vector3D) *GameObject {
	o.endTime = o.startTime.Add(time.Second * 10)
	o.PosVector = mo.PosVector
	o.MoveVector = MoveVector
	o.borderActionFn = borderActionFn_B2_Disable
	o.accelVector = Vector3D{0, 0, 0}
	o.ObjType = GameObjSuperBullet
	return o
}
func (o *GameObject) MakeHommingBullet(mo *GameObject, targetteamid int64, targetid int64) *GameObject {
	o.endTime = o.startTime.Add(time.Second * 60)
	o.PosVector = mo.PosVector
	o.borderActionFn = borderActionFn_B2_Disable
	o.accelVector = Vector3D{0, 0, 0}

	o.MoveVector = Vector3D{0, 0, 0}
	o.targetObjID = targetid
	o.targetTeamID = targetteamid
	o.moveByTimeFn = moveByTimeFn_homming
	o.ObjType = GameObjHommingBullet
	return o
}

func (o *GameObject) MakeRevolutionDecoObj() *GameObject {
	o.moveByTimeFn = moveByTimeFn_clock
	o.borderActionFn = borderActionFn_None
	o.ObjType = GameObjDeco
	return o
}

func (o *GameObject) MakeHomeMarkObj() *GameObject {
	o.PosVector = GameConst.WorldCube.RandVector().Idiv(2)
	o.MoveVector = GameConst.WorldCube.RandVector()
	o.accelVector = GameConst.WorldCube.RandVector()
	o.moveByTimeFn = moveByTimeFn_home
	o.ObjType = GameObjMark
	return o
}

func (o *GameObject) MakeHardObj(pos Vector3D) *GameObject {
	o.PosVector = pos
	o.moveByTimeFn = moveByTimeFn_none
	o.borderActionFn = borderActionFn_None
	o.ObjType = GameObjHard
	return o
}

func (o *GameObject) MakeFoodObj(pos Vector3D) *GameObject {
	o.PosVector = pos
	o.moveByTimeFn = moveByTimeFn_none
	o.borderActionFn = borderActionFn_None
	o.ObjType = GameObjFood
	return o
}
