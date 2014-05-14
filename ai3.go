package go4game

import (
	// "log"
	//"time"
	"math"
	"math/rand"
	"sort"
	"time"
)

// AI3 ----------------------------------------------------------------

type AI3ActionType int

const (
	AI3ActionAccel AI3ActionType = iota
	AI3ActionBullet
	AI3ActionSuperBullet
	AI3ActionHommingBullet
	AI3ActionBurstBullet
	AI3ActionEnd
)

var AI3AP = map[AI3ActionType]int{
	AI3ActionAccel:         GameConst.APAccel,
	AI3ActionBullet:        GameConst.APBullet,
	AI3ActionSuperBullet:   GameConst.APSuperBullet,
	AI3ActionHommingBullet: GameConst.APHommingBullet,
	AI3ActionBurstBullet:   GameConst.APBurstShot,
}

type AI3 struct {
	me              *SPObj
	spp             *SpatialPartition
	worldBound      HyperRect
	ActionPoint     int
	Score           int
	preparedTargets [AI3ActionEnd]AI3AimTargetList
	lastTargets     [AI3ActionEnd]map[int]time.Time
}

func (a *AI3) delOldTagets() {
	var act AI3ActionType

	validold := time.Now().Add(-500 * time.Millisecond)
	act = AI3ActionHommingBullet
	for i, lastfiretime := range a.lastTargets[act] {
		if validold.After(lastfiretime) {
			delete(a.lastTargets[act], i)
		}
	}
	act = AI3ActionSuperBullet
	for i, lastfiretime := range a.lastTargets[act] {
		if validold.After(lastfiretime) {
			delete(a.lastTargets[act], i)
		}
	}

	validold = time.Now().Add(-100 * time.Millisecond)
	act = AI3ActionSuperBullet
	for i, lastfiretime := range a.lastTargets[act] {
		if validold.After(lastfiretime) {
			delete(a.lastTargets[act], i)
		}
	}
}

type AI3AimTargetList []*AI3AimTarget

type AI3AimTarget struct {
	*SPObj
	actFactor [AI3ActionEnd]float64
}

// estmate remain frame to contact( len == 0 )
func (a *AI3) frame2Contact(t *SPObj) float64 {
	collen := a.me.CollisionRadius + t.CollisionRadius
	curlen := a.me.PosVector.LenTo(&t.PosVector) - collen
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

func (a *AI3) CalcEvasionFactor(o *SPObj) float64 {
	// can obj damage me?
	if !InteractionMap[a.me.ObjType][o.ObjType] {
		return -1.0
	}
	// higher is danger : 0 ~ 2
	// anglefactor := 2 - a.me.PosVector.Sub(&o.PosVector).Angle(&o.MoveVector)*2/math.Pi
	anglefactor := 1.0

	typefactor := map[GameObjectType]float64{
		GameObjMain:          2.0,
		GameObjBullet:        1.0,
		GameObjShield:        0.0,
		GameObjHommingBullet: 1.0,
		GameObjSuperBullet:   1.0,
	}[o.ObjType]

	speedrate := ObjDefault.MoveLimit[o.ObjType] / ObjDefault.MoveLimit[a.me.ObjType]

	collen := a.me.CollisionRadius + o.CollisionRadius
	curlen := a.me.PosVector.LenTo(&o.PosVector)
	lenfactor := collen * 5 / curlen

	timefactor := GameConst.FramePerSec / 2 / a.frame2Contact(o) // in 0.5 sec len

	factor := anglefactor * typefactor * lenfactor * timefactor * speedrate
	return factor
}

func (a *AI3) CalcAttackFactor(o *SPObj, bulletType GameObjectType) float64 {
	// is obj attacked by bullet?
	if !InteractionMap[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, estangle := a.calcAims(o, ObjDefault.MoveLimit[bulletType])
	if estpos == nil || !estpos.IsIn(&a.worldBound) { // cannot contact
		return -1.0
	}
	anglefactor := math.Pow(estangle/math.Pi, 2)

	typefactor := map[GameObjectType]float64{
		GameObjMain:          1.2,
		GameObjBullet:        1.0,
		GameObjShield:        0,
		GameObjHommingBullet: 1.3,
		GameObjSuperBullet:   1.4,
	}[o.ObjType]

	collen := a.me.CollisionRadius + o.CollisionRadius
	curlen := a.me.PosVector.LenTo(&o.PosVector)
	lenfactor := collen * 50 / curlen

	factor := anglefactor * typefactor * lenfactor
	return factor
}

func (a *AI3) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			o := AI3AimTarget{
				SPObj: t,
			}
			o.actFactor[AI3ActionAccel] = a.CalcEvasionFactor(t)
			o.actFactor[AI3ActionBullet] = a.CalcAttackFactor(t, GameObjBullet)
			o.actFactor[AI3ActionSuperBullet] = a.CalcAttackFactor(t, GameObjSuperBullet)
			o.actFactor[AI3ActionHommingBullet] = a.CalcAttackFactor(t, GameObjHommingBullet)

			if InteractionMap[a.me.ObjType][t.ObjType] {
				a.preparedTargets[AI3ActionAccel] = append(a.preparedTargets[AI3ActionAccel], &o)
			}
			if InteractionMap[t.ObjType][GameObjBullet] {
				a.preparedTargets[AI3ActionBullet] = append(a.preparedTargets[AI3ActionBullet], &o)
			}
			if t.ObjType == GameObjMain || t.ObjType == GameObjHommingBullet || t.ObjType == GameObjSuperBullet {
				a.preparedTargets[AI3ActionSuperBullet] = append(a.preparedTargets[AI3ActionSuperBullet], &o)
			}
			if t.ObjType == GameObjMain || t.ObjType == GameObjHommingBullet || t.ObjType == GameObjSuperBullet {
				a.preparedTargets[AI3ActionHommingBullet] = append(a.preparedTargets[AI3ActionHommingBullet], &o)
			}
			if InteractionMap[t.ObjType][GameObjBullet] {
				a.preparedTargets[AI3ActionBurstBullet] = append(a.preparedTargets[AI3ActionBurstBullet], &o)
			}
		}
	}
	return false
}

func (a *AI3) calcEvasionVector(t *SPObj) *Vector3D {
	speed := ObjDefault.MoveLimit[a.me.ObjType]
	backvt := a.me.PosVector.Sub(&t.PosVector).NormalizedTo(speed) // backward
	tocentervt := a.me.PosVector.NormalizedTo(speed).Neg()
	return backvt.Add(backvt).Add(tocentervt)
}

func (a *AI3) calcAims(t *SPObj, projectilemovelimit float64) (float64, *Vector3D, float64) {
	dur := a.me.PosVector.CalcAimAheadDur(&t.PosVector, &t.MoveVector, projectilemovelimit)
	if math.IsInf(dur, 1) {
		return math.Inf(1), nil, 0
	}
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estangle := t.MoveVector.Angle(estpos.Sub(&a.me.PosVector))
	return dur, estpos, estangle
}

func (a *AI3) sortActTargets(act AI3ActionType) bool {
	if a.ActionPoint >= AI3AP[act] {
		softFn := func(p1, p2 *AI3AimTarget) bool {
			return p1.actFactor[act] > p2.actFactor[act]
		}
		AI3By(softFn).Sort(a.preparedTargets[act])
		return true
	}
	return false
}

func (a *AI3) MakeAction(packet *GamePacket) *GamePacket {
	if a.lastTargets[0] == nil {
		for act := AI3ActionAccel; act < AI3ActionEnd; act++ {
			a.lastTargets[act] = make(map[int]time.Time)
		}
	}
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score

	if a.spp == nil || a.me == nil {
		return &GamePacket{Cmd: ReqFrameInfo}
	}
	a.worldBound = HyperRect{Min: a.spp.Min, Max: a.spp.Max}
	for i := AI3ActionAccel; i < AI3ActionEnd; i++ {
		a.preparedTargets[i] = make(AI3AimTargetList, 0)
	}
	a.spp.ApplyParts27Fn(a.prepareTarget, a.me.PosVector)

	a.delOldTagets()
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

	var act AI3ActionType

	if act = AI3ActionAccel; a.sortActTargets(act) {
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && rand.Float64() < 0.9 {
				rtn.ClientAct.Accel = a.calcEvasionVector(o.SPObj)
				a.ActionPoint -= AI3AP[act]
				break
			}
		}
		if rtn.ClientAct.Accel == nil && rand.Float64() < 0.5 {
			rtn.ClientAct.Accel = a.me.PosVector.Neg()
			a.ActionPoint -= AI3AP[act]
		}
	}

	if act = AI3ActionSuperBullet; a.sortActTargets(act) {
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && a.lastTargets[act][o.ID].IsZero() {
				_, estpos, _ := a.calcAims(o.SPObj, ObjDefault.MoveLimit[GameObjSuperBullet])
				rtn.ClientAct.SuperBulletMv = estpos.Sub(&a.me.PosVector).NormalizedTo(ObjDefault.MoveLimit[GameObjSuperBullet])
				a.lastTargets[act][o.ID] = time.Now()

				a.ActionPoint -= AI3AP[act]
				break
			}
		}
	}

	if act = AI3ActionHommingBullet; a.sortActTargets(act) {
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && a.lastTargets[act][o.ID].IsZero() {
				rtn.ClientAct.HommingTargetID = []int{o.ID, o.TeamID}
				a.lastTargets[act][o.ID] = time.Now()

				a.ActionPoint -= AI3AP[act]
				break
			}
		}
	}

	if act = AI3ActionBullet; a.sortActTargets(act) {
		for _, o := range a.preparedTargets[act] {
			if o.actFactor[act] > 1 && a.lastTargets[act][o.ID].IsZero() {
				_, estpos, _ := a.calcAims(o.SPObj, ObjDefault.MoveLimit[GameObjBullet])
				rtn.ClientAct.NormalBulletMv = estpos.Sub(&a.me.PosVector).NormalizedTo(ObjDefault.MoveLimit[GameObjBullet])
				a.lastTargets[act][o.ID] = time.Now()

				a.ActionPoint -= AI3AP[act]
				break
			}
		}
	}

	if act = AI3ActionBurstBullet; a.ActionPoint >= AI3AP[act]*(72-len(a.preparedTargets[act])) {
		if a.ActionPoint >= AI3AP[act]*72 {
			rtn.ClientAct.BurstShot = 36
			a.ActionPoint -= AI3AP[act] * rtn.ClientAct.BurstShot
		} else {
			rtn.ClientAct.BurstShot = a.ActionPoint / AI3AP[act] / 2
			a.ActionPoint -= AI3AP[act] * rtn.ClientAct.BurstShot
		}
	}

	return rtn
}

// AI3By is the type of a "less" function that defines the ordering of its AI3AimTarget arguments.
type AI3By func(p1, p2 *AI3AimTarget) bool

// Sort is a method on the function type, AI3By, that sorts the argument slice according to the function.
func (by AI3By) Sort(aimtargets AI3AimTargetList) {
	ps := &AI3AimTargetSorter{
		aimtargets: aimtargets,
		by:         by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// AI3AimTargetSorter joins a AI3By function and a slice of AI3AimTargets to be sorted.
type AI3AimTargetSorter struct {
	aimtargets AI3AimTargetList
	by         func(p1, p2 *AI3AimTarget) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *AI3AimTargetSorter) Len() int {
	return len(s.aimtargets)
}

// Swap is part of sort.Interface.
func (s *AI3AimTargetSorter) Swap(i, j int) {
	s.aimtargets[i], s.aimtargets[j] = s.aimtargets[j], s.aimtargets[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *AI3AimTargetSorter) Less(i, j int) bool {
	return s.by(s.aimtargets[i], s.aimtargets[j])
}