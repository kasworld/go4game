package go4game

import (
	//"log"
	//"time"
	//"math"
	"fmt"
	"math/rand"
	//"sort"
)

type AIBase struct {
	act [ActionEnd]int

	send        *ReqGamePacket
	me          *SPObj
	ActionPoint int
	Score       int
	HomePos     Vector3D
}

func NewAIBase(act [ActionEnd]int) AIActor {
	a := AIBase{act: act}
	return &a
}

func (a AIBase) String() string {
	return fmt.Sprintf("AIBase%v ", a.act)
}

func (a *AIBase) TryAct(act ClientActionType, actnum int) bool {
	if a.act[act] == actnum && a.ActionPoint >= GameConst.AP[act] {
		a.ActionPoint -= GameConst.AP[act]
		return true
	}
	return false
}

func (a *AIBase) CheckAct(act ClientActionType, actnum int, actcount int) bool {
	if a.act[act] == actnum && a.ActionPoint >= GameConst.AP[act]*actcount {
		return true
	}
	return false
}

func (a *AIBase) prepareAI(packet *RspGamePacket) {
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	a.HomePos = packet.TeamInfo.HomePos
	a.send = &ReqGamePacket{
		Cmd: ReqNearInfo,
		ClientAct: &ClientActionPacket{
			Accel:           nil,
			NormalBulletMv:  nil,
			BurstShot:       0,
			HommingTargetID: nil,
			SuperBulletMv:   nil,
		},
	}
}

func (a *AIBase) MakeAction(packet *RspGamePacket) *ReqGamePacket {
	a.prepareAI(packet)
	if a.me == nil {
		return a.send
	}

	var act ClientActionType
	if act = ActionSuperBullet; rand.Float64() < 0.5 && a.TryAct(act, 1) {
		vt := GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjSuperBullet])
		a.send.ClientAct.SuperBulletMv = &vt
	}
	if act = ActionHommingBullet; rand.Float64() < 0.5 && a.TryAct(act, 1) {
		a.send.ClientAct.HommingTargetID = IDList{a.me.ID, a.me.TeamID}
	}
	if act = ActionBullet; rand.Float64() < 0.5 && a.TryAct(act, 1) {
		vt := GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjBullet])
		a.send.ClientAct.NormalBulletMv = &vt
	}
	if act = ActionAccel; a.TryAct(act, 1) {
		var vt Vector3D
		if rand.Float64() < 0.5 {
			vt = GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjMain])
		} else {
			vt = a.HomePos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjMain])
		}
		a.send.ClientAct.Accel = &vt
	}
	if act = ActionAccel; a.TryAct(act, 2) {
		vt := a.me.MoveVector.Neg()
		a.send.ClientAct.Accel = &vt
	}
	if act = ActionBurstBullet; rand.Float64() < 0.5 && a.CheckAct(act, 1, 40) {
		a.send.ClientAct.BurstShot = a.ActionPoint/GameConst.AP[act] - 4
		a.ActionPoint -= GameConst.AP[act] * a.send.ClientAct.BurstShot
	}
	return a.send
}
