package go4game

import (
	//"log"
	"fmt"
	//"math"
	//"math/rand"
	"sort"
	"time"
)

type FactorCalcFn func(a *AIAdv, o *AIAdvAimTarget) float64
type VectorMakeFn func(a *AIAdv) *Vector3D
type IDListMakeFn func(a *AIAdv) IDList
type IntMakeFn func(a *AIAdv) int

type AIVector3DAct struct {
	CalcFn FactorCalcFn
	Fn     VectorMakeFn
}
type AIIDListAct struct {
	CalcFn FactorCalcFn
	Fn     IDListMakeFn
}
type AIIntAct struct {
	CalcFn IntMakeFn
	Fn     IntMakeFn
}

type AIAdvFns struct {
	SuperFns   []AIVector3DAct
	BulletFns  []AIVector3DAct
	AccelFns   []AIVector3DAct
	HommingFns []AIIDListAct
	BurstFns   []AIIntAct
}

type AIAdv struct {
	AIAdvFns
	act  [ActionEnd]int
	name string

	send            *ReqGamePacket
	me              *SPObj
	ActionPoint     int
	Score           int
	HomePos         Vector3D
	preparedTargets [ActionEnd]AIAdvAimTargetList
	lastTargets     [ActionEnd]map[int64]time.Time
}

func (a AIAdv) String() string {
	return fmt.Sprintf("AI%v%v ", a.name, a.act)
}

func (a *AIAdv) UseAP(act ClientActionType) bool {
	return a.UseAPn(act, 1)
}

func (a *AIAdv) UseAPn(act ClientActionType, actcount int) bool {
	if a.CheckActn(act, actcount) {
		a.ActionPoint -= GameConst.AP[act] * actcount
		return true
	}
	return false
}

func (a *AIAdv) CheckActn(act ClientActionType, actcount int) bool {
	return a.ActionPoint >= GameConst.AP[act]*actcount
}

func (a *AIAdv) CheckAct(act ClientActionType) bool {
	return a.CheckActn(act, 1)
}

func (a *AIAdv) prepareAI(packet *RspGamePacket) {
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	a.HomePos = packet.TeamInfo.HomePos
	for i := ActionAccel; i < ActionEnd; i++ {
		a.preparedTargets[i] = make(AIAdvAimTargetList, 0)
	}
	a.delOldTagets()
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

func (a *AIAdv) getATAIN(act ClientActionType) (ClientActionType, int) {
	return act, a.act[act]
}

func (a *AIAdv) checkSort(act ClientActionType, calcfn FactorCalcFn) bool {
	if !a.CheckAct(act) {
		return false
	}
	if calcfn != nil {
		for _, o := range a.preparedTargets[act] {
			o.actFactor = calcfn(a, o)
		}
		sort.Sort(a.preparedTargets[act])
	}
	return true
}

func (a *AIAdv) MakeAction(packet *RspGamePacket) *ReqGamePacket {
	a.prepareAI(packet)
	if a.me == nil {
		return a.send
	}
	var act ClientActionType
	for _, t := range packet.NearObjs {
		if a.me.TeamID == t.TeamID {
			continue
		}
		o := AIAdvAimTarget{SPObj: t}
		if act = ActionAccel; GameConst.IsInteract[a.me.ObjType][t.ObjType] {
			a.preparedTargets[act] = append(a.preparedTargets[act], &o)
		}
		if act = ActionBullet; GameConst.IsInteract[t.ObjType][GameObjBullet] {
			a.preparedTargets[act] = append(a.preparedTargets[act], &o)
		}
		if act = ActionSuperBullet; GameConst.IsInteract[t.ObjType][GameObjSuperBullet] {
			a.preparedTargets[act] = append(a.preparedTargets[act], &o)
		}
		if act = ActionHommingBullet; GameConst.IsInteract[t.ObjType][GameObjHommingBullet] {
			a.preparedTargets[act] = append(a.preparedTargets[act], &o)
		}
		if act = ActionBurstBullet; GameConst.IsInteract[t.ObjType][GameObjBullet] {
			a.preparedTargets[act] = append(a.preparedTargets[act], &o)
		}
	}

	if act, actnum := a.getATAIN(ActionSuperBullet); a.SuperFns[actnum].Fn != nil && a.checkSort(act, a.SuperFns[actnum].CalcFn) {
		a.send.ClientAct.SuperBulletMv = a.SuperFns[actnum].Fn(a)
		if a.send.ClientAct.SuperBulletMv != nil {
			a.UseAP(act)
		}
	}
	if act, actnum := a.getATAIN(ActionHommingBullet); a.HommingFns[actnum].Fn != nil && a.checkSort(act, a.HommingFns[actnum].CalcFn) {
		a.send.ClientAct.HommingTargetID = a.HommingFns[actnum].Fn(a)
		if a.send.ClientAct.HommingTargetID != nil {
			a.UseAP(act)
		}
	}
	if act, actnum := a.getATAIN(ActionBullet); a.BulletFns[actnum].Fn != nil && a.checkSort(act, a.BulletFns[actnum].CalcFn) {
		a.send.ClientAct.NormalBulletMv = a.BulletFns[actnum].Fn(a)
		if a.send.ClientAct.NormalBulletMv != nil {
			a.UseAP(act)
		}
	}
	if act, actnum := a.getATAIN(ActionAccel); a.AccelFns[actnum].Fn != nil && a.checkSort(act, a.AccelFns[actnum].CalcFn) {
		a.send.ClientAct.Accel = a.AccelFns[actnum].Fn(a)
		if a.send.ClientAct.Accel != nil {
			a.UseAP(act)
		}
	}
	if act, actnum := a.getATAIN(ActionBurstBullet); a.BurstFns[actnum].Fn != nil && a.CheckActn(act, a.BurstFns[actnum].CalcFn(a)) {
		if a.BurstFns[actnum].Fn != nil {
			a.send.ClientAct.BurstShot = a.BurstFns[actnum].Fn(a)
			a.UseAPn(act, a.send.ClientAct.BurstShot)
		}
	}
	return a.send
}

func (a *AIAdv) delOldTagets() {
	var act ClientActionType

	validold := time.Now().Add(-500 * time.Millisecond)
	act = ActionHommingBullet
	for i, lastfiretime := range a.lastTargets[act] {
		if validold.After(lastfiretime) {
			delete(a.lastTargets[act], i)
		}
	}
	act = ActionSuperBullet
	for i, lastfiretime := range a.lastTargets[act] {
		if validold.After(lastfiretime) {
			delete(a.lastTargets[act], i)
		}
	}

	validold = time.Now().Add(-100 * time.Millisecond)
	act = ActionBullet
	for i, lastfiretime := range a.lastTargets[act] {
		if validold.After(lastfiretime) {
			delete(a.lastTargets[act], i)
		}
	}
}

type AIAdvAimTargetList []*AIAdvAimTarget
type AIAdvAimTarget struct {
	*SPObj
	actFactor float64
}

func (s AIAdvAimTargetList) Len() int {
	return len(s)
}
func (s AIAdvAimTargetList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s AIAdvAimTargetList) Less(i, j int) bool {
	return s[i].actFactor > s[j].actFactor
}
