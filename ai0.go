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

	if a.ActionPoint >= ActionPoints[ActionSuperBullet] && rand.Float64() < 0.5 {
		tmp := RandVector(a.spp.Min, a.spp.Max)
		rtn.ClientAct.SuperBulletMv = &tmp
		a.ActionPoint -= ActionPoints[ActionSuperBullet]
	}

	if a.ActionPoint >= ActionPoints[ActionHommingBullet] && rand.Float64() < 0.5 {
		rtn.ClientAct.HommingTargetID = IDList{a.me.ID, a.me.TeamID}
		a.ActionPoint -= ActionPoints[ActionHommingBullet]
	}

	if a.ActionPoint >= ActionPoints[ActionBullet] && rand.Float64() < 0.5 {
		tmp := RandVector(a.spp.Min, a.spp.Max)
		rtn.ClientAct.NormalBulletMv = &tmp
		a.ActionPoint -= ActionPoints[ActionBullet]
	}

	if a.ActionPoint >= ActionPoints[ActionAccel] {
		if rand.Float64() < 0.5 {
			tmp := RandVector(a.spp.Min, a.spp.Max)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= ActionPoints[ActionAccel]
		} else {
			tmp := a.HomePos.Sub(a.me.PosVector)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= ActionPoints[ActionAccel]
		}
	}

	if a.ActionPoint >= ActionPoints[ActionBurstBullet]*40 && rand.Float64() < 0.5 {
		rtn.ClientAct.BurstShot = a.ActionPoint/ActionPoints[ActionBurstBullet] - 4
		a.ActionPoint -= ActionPoints[ActionBurstBullet] * rtn.ClientAct.BurstShot
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

	if a.ActionPoint >= ActionPoints[ActionHommingBullet] && rand.Float64() < 0.5 {
		rtn.ClientAct.HommingTargetID = IDList{a.me.ID, a.me.TeamID}
		a.ActionPoint -= ActionPoints[ActionHommingBullet]
	}

	if a.ActionPoint >= ActionPoints[ActionAccel] {
		if rand.Float64() < 0.5 {
			tmp := RandVector(a.spp.Min, a.spp.Max)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= ActionPoints[ActionAccel]
		} else {
			tmp := a.HomePos.Sub(a.me.PosVector)
			rtn.ClientAct.Accel = &tmp
			a.ActionPoint -= ActionPoints[ActionAccel]
		}
	}

	return rtn
}
