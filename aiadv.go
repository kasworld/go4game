package go4game

import (
	//"log"
	"fmt"
	//"math"
	//"math/rand"
	"sort"
	"time"

)

type AIAdvFns struct {
    CalcSuperFactorFn   []func(a *AIAdv, o *AIAdvAimTarget) float64
    SuperBulletFn       []func(a *AIAdv) *Vector3D
    CalcBulletFactorFn  []func(a *AIAdv, o *AIAdvAimTarget) float64
    NormalBulletFn      []func(a *AIAdv) *Vector3D
    CalcHommingFactorFn []func(a *AIAdv, o *AIAdvAimTarget) float64
    HommingBulletFn     []func(a *AIAdv) IDList
    CalcAccelFactorFn   []func(a *AIAdv, o *AIAdvAimTarget) float64
    AccelFn             []func(a *AIAdv) *Vector3D
    CalcBurstFactorFn   []func(a *AIAdv) int
    BurstBulletFn       []func(a *AIAdv) int
}

type AIAdv struct {
    AIAdvFns
	act                 [ActionEnd]int

	send            *ReqGamePacket
	me              *SPObj
	ActionPoint     int
	Score           int
	HomePos         Vector3D
	preparedTargets [ActionEnd]AIAdvAimTargetList
	lastTargets     [ActionEnd]map[int64]time.Time
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

func (a *AIAdv) checkSort(act ClientActionType, calcfn []func(*AIAdv, *AIAdvAimTarget) float64) (ClientActionType, int, bool) {
	actnum := a.act[act]
	if a.CheckAct(act) && calcfn[actnum] != nil {
		for _, o := range a.preparedTargets[act] {
			o.actFactor = calcfn[actnum](a, o)
		}
		sort.Sort(a.preparedTargets[act])
		return act, actnum, true
	} else {
		return act, actnum, false
	}
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

	if act, actnum, ok := a.checkSort(ActionSuperBullet, a.CalcSuperFactorFn); ok {
		if a.SuperBulletFn[actnum] != nil {
			a.send.ClientAct.SuperBulletMv = a.SuperBulletFn[actnum](a)
		}
		if a.send.ClientAct.SuperBulletMv != nil {
			a.UseAP(act)
		}
	}
	if act, actnum, ok := a.checkSort(ActionHommingBullet, a.CalcHommingFactorFn); ok {
		if a.HommingBulletFn[actnum] != nil {
			a.send.ClientAct.HommingTargetID = a.HommingBulletFn[actnum](a)
		}
		if a.send.ClientAct.HommingTargetID != nil {
			a.UseAP(act)
		}
	}
	if act, actnum, ok := a.checkSort(ActionBullet, a.CalcBulletFactorFn); ok {
		if a.NormalBulletFn[actnum] != nil {
			a.send.ClientAct.NormalBulletMv = a.NormalBulletFn[actnum](a)
		}
		if a.send.ClientAct.NormalBulletMv != nil {
			a.UseAP(act)
		}
	}
	if act, actnum, ok := a.checkSort(ActionAccel, a.CalcAccelFactorFn); ok {
		if a.AccelFn[actnum] != nil {
			a.send.ClientAct.Accel = a.AccelFn[actnum](a)
		}
		if a.send.ClientAct.Accel != nil {
			a.UseAP(act)
		}
	}
	if act, actnum := a.getATAIN(ActionBurstBullet); a.CheckActn(act, a.CalcBurstFactorFn[actnum](a) ) {
		if a.BurstBulletFn[actnum] != nil {
			a.send.ClientAct.BurstShot = a.BurstBulletFn[actnum](a)
		}
		a.UseAPn(act, a.send.ClientAct.BurstShot)
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

func (a AIAdv) String() string {
	return fmt.Sprintf("AIAdv%v ", a.act)
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
	return s[i].actFactor < s[j].actFactor
}
