package go4game

import (
//"log"
//"time"
)

type AIConn struct {
	//pteam          *Team
	ppos           [3]int
	bulletargetpos *Vector3D
	me             SPObj
	spp            *SpatialPartition
}

func (a *AIConn) SelectTarget(s SPObjList) bool {
	for _, o := range s {
		teamrule := o.TeamID != a.me.TeamID
		if teamrule {
			a.bulletargetpos = &o.PosVector
			return true
		}
	}
	return false
}

func (a *AIConn) makeAIAction() *GamePacket {
	if a.spp == nil {
		return &GamePacket{Cmd: ReqAIAct}
	}
	a.ppos = a.spp.Pos2PartPos(a.me.PosVector)
	a.bulletargetpos = nil
	a.spp.ApplyPartsFn(a.SelectTarget, a.me.PosVector, a.spp.MaxObjectRadius)

	var bulletTargetPos *Vector3D
	if a.bulletargetpos != nil {
		bulletTargetPos = a.bulletargetpos.Sub(&a.me.PosVector).Normalized().Imul(500)
	}

	return &GamePacket{
		Cmd: ReqAIAct,
		ClientAct: &ClientActionPacket{
			Accel:          RandVector3D(-500, 500),
			NormalBulletMv: bulletTargetPos,
		},
	}
}
