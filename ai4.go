package go4game

import (
	//"log"
	//"time"
	"math"
	//"math/rand"
	"sort"
	"time"
)

// AI4 ----------------------------------------------------------------

type AI4 struct {
	me              *SPObj
	ActionPoint     int
	Score           int
	HomePos         Vector3D
	preparedTargets [ActionEnd]AI4AimTargetList
	lastTargets     [ActionEnd]map[int64]time.Time
}

// from gameobj moveByTimeFn_accel
func (m *SPObj) TestMoveByAccel(accelVector Vector3D) Vector3D {
	dur := 1000 / GameConst.FramePerSec
	MoveVector := m.MoveVector.Add(accelVector.Imul(dur))
	if MoveVector.Abs() > GameConst.MoveLimit[m.ObjType] {
		MoveVector = MoveVector.NormalizedTo(GameConst.MoveLimit[m.ObjType])
	}
	PosVector := m.PosVector.Add(MoveVector.Imul(dur))
	return PosVector
}

func NewAI4() AIActor {
	return &AI4{}
}

func (a *AI4) delOldTagets() {
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

type AI4AimTargetList []*AI4AimTarget

type AI4AimTarget struct {
	*SPObj
	actFactor [ActionEnd]float64
}

// estmate remain frame to contact( len == 0 )
func (a *AI4) frame2Contact(t *SPObj) float64 {
	collen := math.Sqrt(GameConst.ObjSqd[a.me.ObjType][t.ObjType])
	curlen := a.me.PosVector.LenTo(t.PosVector) - collen
	nextposme := a.me.PosVector.Add(a.me.MoveVector.Idiv(GameConst.FramePerSec))
	nextpost := t.PosVector.Add(t.MoveVector.Idiv(GameConst.FramePerSec))
	nextlen := nextposme.LenTo(nextpost) - collen
	if curlen <= 0 || nextlen <= 0 {
		return 0
	}
	changelenperframe := nextlen - curlen // + farer , - nearer
	if changelenperframe >= 0 {
		return math.Inf(1) // inf frame to contact
	}
	return curlen / -changelenperframe
}

func (a *AI4) calcEvasionVector(t *SPObj) *Vector3D {
	speed := GameConst.MoveLimit[a.me.ObjType]
	backvt := a.me.PosVector.Sub(t.PosVector).NormalizedTo(speed) // backward
	tohomevt := a.HomePos.Sub(a.me.PosVector).NormalizedTo(speed) // to home pos
	rtn := backvt.Add(backvt).Add(tohomevt)
	return &rtn
}

func (a *AI4) calcLenRate(t *SPObj) float64 {
	collen := GameConst.Radius[a.me.ObjType] + GameConst.Radius[t.ObjType]
	curlen := a.me.PosVector.LenTo(t.PosVector) - collen

	nextposme := a.me.PosVector.Add(a.me.MoveVector.Idiv(GameConst.FramePerSec))
	nextpost := t.PosVector.Add(t.MoveVector.Idiv(GameConst.FramePerSec))

	nextlen := nextposme.LenTo(nextpost) - collen
	if curlen <= 0 || nextlen <= 0 {
		return GameConst.WorldDiag // math.Inf(1)
	} else {
		return curlen / nextlen
	}
}

func (a *AI4) CalcEvasionFactor(o *SPObj) float64 {
	// can obj damage me?
	if !GameConst.IsInteract[a.me.ObjType][o.ObjType] {
		return -1.0
	}
	// higher is danger : 0 ~ 2
	// anglefactor := 2 - a.me.PosVector.Sub(&o.PosVector).Angle(&o.MoveVector)*2/math.Pi
	// anglefactor := 1.0

	// typefactor := [GameObjEnd]float64{
	// 	GameObjMain:          2.0,
	// 	GameObjBullet:        1.0,
	// 	GameObjShield:        0.0,
	// 	GameObjHommingBullet: 1.0,
	// 	GameObjSuperBullet:   1.0,
	// 	GameObjHard:          1.0,
	// }[o.ObjType]

	//speedrate := GameConst.MoveLimit[o.ObjType] / GameConst.MoveLimit[a.me.ObjType]

	lenfactor := a.calcLenRate(o)
	return lenfactor

	// timefactor := GameConst.FramePerSec / 2 / a.frame2Contact(o) // in 0.5 sec len

	// factor := anglefactor * typefactor * lenfactor * timefactor //* speedrate
	// return factor
}

func (a *AI4) CalcAttackFactor(o *SPObj, bulletType GameObjectType) float64 {
	// is obj attacked by bullet?
	if !GameConst.IsInteract[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, estangle := a.calcAims(o, GameConst.MoveLimit[bulletType])
	if estpos == nil || !estpos.IsIn(GameConst.WorldCube) { // cannot contact
		return -1.0
	}
	anglefactor := math.Pow(estangle/math.Pi, 2)

	typefactor := [GameObjEnd]float64{
		GameObjMain:          1.2,
		GameObjBullet:        1.0,
		GameObjShield:        0,
		GameObjHommingBullet: 1.3,
		GameObjSuperBullet:   1.4,
	}[o.ObjType]

	collen := math.Sqrt(GameConst.ObjSqd[a.me.ObjType][o.ObjType])
	curlen := a.me.PosVector.LenTo(o.PosVector)
	lenfactor := collen * 50 / curlen

	factor := anglefactor * typefactor * lenfactor
	return factor
}

func (a *AI4) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			o := AI4AimTarget{
				SPObj: t,
			}
			o.actFactor[ActionAccel] = a.CalcEvasionFactor(t)
			o.actFactor[ActionBullet] = a.CalcAttackFactor(t, GameObjBullet)
			o.actFactor[ActionSuperBullet] = a.CalcAttackFactor(t, GameObjSuperBullet)
			o.actFactor[ActionHommingBullet] = a.CalcAttackFactor(t, GameObjHommingBullet)

			if GameConst.IsInteract[a.me.ObjType][t.ObjType] {
				a.preparedTargets[ActionAccel] = append(a.preparedTargets[ActionAccel], &o)
			}
			if GameConst.IsInteract[t.ObjType][GameObjBullet] {
				a.preparedTargets[ActionBullet] = append(a.preparedTargets[ActionBullet], &o)
			}
			if GameConst.IsInteract[t.ObjType][GameObjSuperBullet] {
				a.preparedTargets[ActionSuperBullet] = append(a.preparedTargets[ActionSuperBullet], &o)
			}
			if GameConst.IsInteract[t.ObjType][GameObjHommingBullet] {
				a.preparedTargets[ActionHommingBullet] = append(a.preparedTargets[ActionHommingBullet], &o)
			}
			if GameConst.IsInteract[t.ObjType][GameObjBullet] {
				a.preparedTargets[ActionBurstBullet] = append(a.preparedTargets[ActionBurstBullet], &o)
			}
		}
	}
	return false
}

func (a *AI4) calcAims(t *SPObj, projectilemovelimit float64) (float64, *Vector3D, float64) {
	dur := a.me.PosVector.CalcAimAheadDur(t.PosVector, t.MoveVector, projectilemovelimit)
	if math.IsInf(dur, 1) {
		return math.Inf(1), nil, 0
	}
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estangle := t.MoveVector.Angle(estpos.Sub(a.me.PosVector))
	return dur, &estpos, estangle
}

func (a *AI4) sortActTargets(act ClientActionType) bool {
	if a.ActionPoint >= GameConst.AP[act] {
		softFn := func(p1, p2 *AI4AimTarget) bool {
			return p1.actFactor[act] > p2.actFactor[act]
		}
		AI4By(softFn).Sort(a.preparedTargets[act])
		return true
	}
	return false
}

func (a *AI4) calcBackVector(t *SPObj, evfactor float64) Vector3D {
	speed := GameConst.MoveLimit[a.me.ObjType]
	return a.me.PosVector.Sub(t.PosVector).NormalizedTo(evfactor * speed)
}

func (a *AI4) prepareAI(packet *RspGamePacket) *ReqGamePacket {
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	a.HomePos = packet.TeamInfo.HomePos
	for i := ActionAccel; i < ActionEnd; i++ {
		a.preparedTargets[i] = make(AI4AimTargetList, 0)
	}
	a.prepareTarget(packet.NearObjs)
	a.delOldTagets()
	return &ReqGamePacket{
		Cmd: ReqNearInfo,
		ClientAct: &ClientActionPacket{
			Accel:           &Vector3D{},
			NormalBulletMv:  nil,
			BurstShot:       0,
			HommingTargetID: nil,
			SuperBulletMv:   nil,
		},
	}

}

func (a *AI4) MakeAction(packet *RspGamePacket) *ReqGamePacket {
	if a.lastTargets[0] == nil {
		//log.Printf("init historydata ")
		for act := ActionAccel; act < ActionEnd; act++ {
			a.lastTargets[act] = make(map[int64]time.Time)
		}
	}
	if packet.TeamInfo.SPObj == nil {
		return &ReqGamePacket{Cmd: ReqNearInfo}
	}
	rtn := a.prepareAI(packet)
	var act ClientActionType

	if act = ActionAccel; a.sortActTargets(act) {
		a.ActionPoint -= GameConst.AP[act]
		for i, o := range a.preparedTargets[act] {
			if i > 3 { // apply max 3 target
				break
			}
			if o.actFactor[act] > 1 {
				tmp := rtn.ClientAct.Accel.Add(a.calcBackVector(o.SPObj, o.actFactor[act]))
				rtn.ClientAct.Accel = &tmp
				//log.Printf("accel %v %v %v ", i, o.actFactor[act], rtn.ClientAct.Accel)
			}
		}
		if rtn.ClientAct.Accel.Abs() < 10 {
			tmp := rtn.ClientAct.Accel.Add(a.HomePos.Sub(a.me.PosVector))
			rtn.ClientAct.Accel = &tmp
			//log.Printf("accel to home %v ", rtn.ClientAct.Accel)
		}
	}

	if act = ActionSuperBullet; a.sortActTargets(act) {
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && a.lastTargets[act][o.ID].IsZero() {
				_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[GameObjSuperBullet])
				tmp := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjSuperBullet])
				rtn.ClientAct.SuperBulletMv = &tmp
				a.lastTargets[act][o.ID] = time.Now()

				a.ActionPoint -= GameConst.AP[act]
				break
			}
		}
	}

	if act = ActionHommingBullet; a.sortActTargets(act) {
		// offencive homming
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && a.lastTargets[act][o.ID].IsZero() {
				rtn.ClientAct.HommingTargetID = IDList{o.ID, o.TeamID}
				a.lastTargets[act][o.ID] = time.Now()

				a.ActionPoint -= GameConst.AP[act]
				break
			}
		}
		// defencive homming
		if rtn.ClientAct.HommingTargetID == nil {
			o := a.me
			if a.lastTargets[act][o.ID].IsZero() {
				rtn.ClientAct.HommingTargetID = IDList{o.ID, o.TeamID}
				a.lastTargets[act][o.ID] = time.Now()
				a.ActionPoint -= GameConst.AP[act]
			}
		}
	}

	if act = ActionBullet; a.sortActTargets(act) {
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && a.lastTargets[act][o.ID].IsZero() {
				_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[GameObjBullet])
				tmp := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjBullet])
				rtn.ClientAct.NormalBulletMv = &tmp
				a.lastTargets[act][o.ID] = time.Now()

				a.ActionPoint -= GameConst.AP[act]
				break
			}
		}
	}

	if act = ActionBurstBullet; a.ActionPoint >= GameConst.AP[act]*(72-len(a.preparedTargets[act])) {
		if a.ActionPoint >= GameConst.AP[act]*72 {
			rtn.ClientAct.BurstShot = 36
			a.ActionPoint -= GameConst.AP[act] * rtn.ClientAct.BurstShot
		} else {
			rtn.ClientAct.BurstShot = a.ActionPoint / GameConst.AP[act] / 2
			a.ActionPoint -= GameConst.AP[act] * rtn.ClientAct.BurstShot
		}
	}

	return rtn
}

// AI4By is the type of a "less" function that defines the ordering of its AI4AimTarget arguments.
type AI4By func(p1, p2 *AI4AimTarget) bool

// Sort is a method on the function type, AI4By, that sorts the argument slice according to the function.
func (by AI4By) Sort(aimtargets AI4AimTargetList) {
	ps := &AI4AimTargetSorter{
		aimtargets: aimtargets,
		by:         by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// AI4AimTargetSorter joins a AI4By function and a slice of AI4AimTargets to be sorted.
type AI4AimTargetSorter struct {
	aimtargets AI4AimTargetList
	by         func(p1, p2 *AI4AimTarget) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *AI4AimTargetSorter) Len() int {
	return len(s.aimtargets)
}

// Swap is part of sort.Interface.
func (s *AI4AimTargetSorter) Swap(i, j int) {
	s.aimtargets[i], s.aimtargets[j] = s.aimtargets[j], s.aimtargets[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *AI4AimTargetSorter) Less(i, j int) bool {
	return s.by(s.aimtargets[i], s.aimtargets[j])
}
