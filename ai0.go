package go4game

import (
	//"log"
	//"time"
	//"math"
	"math/rand"
	//"sort"
)

// AINothing ----------------------------------------------------------------
type AINothing struct {
}

func (a *AINothing) MakeAction(packet *GamePacket) *GamePacket {
	var bulletMoveVector *Vector3D = nil
	var accvt *Vector3D = nil
	var burstCount int = 0
	var hommingTargetID IDList // objid, teamid
	var superBulletMv *Vector3D = nil
	return &GamePacket{
		Cmd: ReqFrameInfo,
		ClientAct: &ClientActionPacket{
			Accel:           accvt,
			NormalBulletMv:  bulletMoveVector,
			BurstShot:       burstCount,
			HommingTargetID: hommingTargetID,
			SuperBulletMv:   superBulletMv,
		},
	}
}

// AIRandom ----------------------------------------------------------------
type AIRandom struct {
	me          *SPObj
	spp         *SpatialPartition
	worldBound  *HyperRect
	ActionPoint int
	Score       int
	HomePos     Vector3D
}

func (a *AIRandom) MakeAction(packet *GamePacket) *GamePacket {
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	a.HomePos = packet.TeamInfo.HomePos

	if a.spp == nil || a.me == nil {
		return &GamePacket{Cmd: ReqFrameInfo}
	}
	a.worldBound = &HyperRect{Min: a.spp.Min, Max: a.spp.Max}

	rtn := &GamePacket{
		Cmd: ReqFrameInfo,
		ClientAct: &ClientActionPacket{
			Accel:           nil,
			NormalBulletMv:  nil,
			BurstShot:       0,
			HommingTargetID: nil,
			SuperBulletMv:   nil,
		},
	}

	if a.ActionPoint >= GameConst.AP[ActionSuperBullet] && rand.Float64() < 0.5 {
		tmp := RandVector(a.spp.Min, a.spp.Max)
		rtn.ClientAct.SuperBulletMv = &tmp
		a.ActionPoint -= GameConst.AP[ActionSuperBullet]
	}

	if a.ActionPoint >= GameConst.AP[ActionHommingBullet] && rand.Float64() < 0.5 {
		rtn.ClientAct.HommingTargetID = IDList{a.me.ID, a.me.TeamID}
		a.ActionPoint -= GameConst.AP[ActionHommingBullet]
	}

	if a.ActionPoint >= GameConst.AP[ActionBullet] && rand.Float64() < 0.5 {
		tmp := RandVector(a.spp.Min, a.spp.Max)
		rtn.ClientAct.NormalBulletMv = &tmp
		a.ActionPoint -= GameConst.AP[ActionBullet]
	}

	if a.ActionPoint >= GameConst.AP[ActionAccel] {
		if rand.Float64() < 0.5 {
			tmp := RandVector(a.spp.Min, a.spp.Max)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= GameConst.AP[ActionAccel]
		} else {
			tmp := a.HomePos.Sub(a.me.PosVector)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= GameConst.AP[ActionAccel]
		}
	}

	if a.ActionPoint >= GameConst.AP[ActionBurstBullet]*40 && rand.Float64() < 0.5 {
		rtn.ClientAct.BurstShot = a.ActionPoint/GameConst.AP[ActionBurstBullet] - 4
		a.ActionPoint -= GameConst.AP[ActionBurstBullet] * rtn.ClientAct.BurstShot
	}

	return rtn
}

// AICloud ----------------------------------------------------------------
type AICloud struct {
	me          *SPObj
	spp         *SpatialPartition
	worldBound  *HyperRect
	ActionPoint int
	Score       int
	HomePos     Vector3D
}

func (a *AICloud) MakeAction(packet *GamePacket) *GamePacket {
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	a.HomePos = packet.TeamInfo.HomePos

	if a.spp == nil || a.me == nil {
		return &GamePacket{Cmd: ReqFrameInfo}
	}
	a.worldBound = &HyperRect{Min: a.spp.Min, Max: a.spp.Max}

	rtn := &GamePacket{
		Cmd: ReqFrameInfo,
		ClientAct: &ClientActionPacket{
			Accel:           nil,
			NormalBulletMv:  nil,
			BurstShot:       0,
			HommingTargetID: nil,
			SuperBulletMv:   nil,
		},
	}

	if a.ActionPoint >= GameConst.AP[ActionHommingBullet] && rand.Float64() < 0.5 {
		rtn.ClientAct.HommingTargetID = IDList{a.me.ID, a.me.TeamID}
		a.ActionPoint -= GameConst.AP[ActionHommingBullet]
	}

	if a.ActionPoint >= GameConst.AP[ActionAccel] {
		if rand.Float64() < 0.5 {
			tmp := RandVector(a.spp.Min, a.spp.Max)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= GameConst.AP[ActionAccel]
		} else {
			tmp := a.HomePos.Sub(a.me.PosVector)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= GameConst.AP[ActionAccel]
		}
	}

	return rtn
}
