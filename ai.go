package go4game

import (
//"log"
//"time"
)

type AIConn struct {
	pteam          *Team
	ppos           [3]int
	bulletargetpos *Vector3D
}

func (a *AIConn) SelectTarget(s *SPObj, me *GameObject) bool {
	teamrule := s.TeamID != me.PTeam.ID
	if teamrule {
		a.bulletargetpos = &s.PosVector
		return true
	}
	return false
}

func (a *AIConn) makeAIAction(spp *SpatialPartition) *GamePacket {
	me := a.pteam.findMainObj()
	if me == nil {
		return &GamePacket{Cmd: ReqAIAct}
	}
	a.ppos = spp.GetPartPos(me.PosVector)
	a.bulletargetpos = nil
	spp.ApplyCollisionAction3(a.SelectTarget, me)

	var bulletTargetPos *Vector3D
	if a.bulletargetpos != nil {
		bulletTargetPos = a.bulletargetpos.Sub(&me.PosVector).Normalized().Imul(500)
	}

	return &GamePacket{
		Cmd: ReqAIAct,
		ClientAct: &ClientActionPacket{
			Accel:          RandVector3D(-500, 500),
			NormalBulletMv: bulletTargetPos,
		},
	}
}
