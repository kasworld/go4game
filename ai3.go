package go4game

import (
	//"log"
	//"time"
	"math"
	"math/rand"
	"sort"
	"time"
)

// AI3 ----------------------------------------------------------------
type AI3 struct {
	me          *SPObj
	spp         *SpatialPartition
	targetlist  AI3AimTargetList
	mainobjlist AI3AimTargetList
	worldBound  HyperRect
	ActionPoint int
	Score       int
	lastHommingTarget map[int]time.Time
}

func (a *AI3) delOld() {
	validold := time.Now().Add(-1*time.Second)
	ts := a.lastHommingTarget
	for i, o := range ts {
		if validold.Before(o) {
			delete(ts, i)
		}
	}
}

type AI3AimTargetList []*AI3AimTarget

type AI3AimTarget struct {
	*SPObj
	EvasionFactor       float64
	BulletAttackFactor  float64
	SBulletAttackFactor float64
	HommingAttackFactor float64
}

// estmate remain frame to contact( len == 0 )
func (a *AI3) frame2Contact(t *SPObj) float64 {
	collen := a.me.CollisionRadius + t.CollisionRadius
	curlen := a.me.PosVector.LenTo(&t.PosVector) - collen
	nextposme := a.me.PosVector.Add(a.me.MoveVector.Idiv(60.0))
	nextpost := t.PosVector.Add(t.MoveVector.Idiv(60.0))
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
	anglefactor := 2 - a.me.PosVector.Sub(&o.PosVector).Angle(&o.MoveVector)*2/math.Pi

	typefactor := map[GameObjectType]float64{
		GameObjMain:          1.2,
		GameObjBullet:        1.0,
		GameObjShield:        1.0,
		GameObjHommingBullet: 1.3,
		GameObjSuperBullet:   1.4,
	}[o.ObjType]

	collen := a.me.CollisionRadius + o.CollisionRadius
	curlen := a.me.PosVector.LenTo(&o.PosVector)
	lenfactor := collen * 10 / curlen

	factor := anglefactor * typefactor * lenfactor
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
		GameObjShield:        1.0,
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
				SPObj:               t,
				EvasionFactor:       a.CalcEvasionFactor(t),
				BulletAttackFactor:  a.CalcAttackFactor(t, GameObjBullet),
				SBulletAttackFactor: a.CalcAttackFactor(t, GameObjSuperBullet),
				HommingAttackFactor: a.CalcAttackFactor(t, GameObjHommingBullet),
			}
			a.targetlist = append(a.targetlist, &o)
			if t.ObjType == GameObjMain {
				a.mainobjlist = append(a.mainobjlist, &o)
			}
		}
	}
	return false
}

func (a *AI3) calcEvasionVector(t *SPObj) *Vector3D {
	speed := (a.me.CollisionRadius + t.CollisionRadius) * 60
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

func (a *AI3) MakeAction(packet *GamePacket) *GamePacket {
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score

	if a.spp == nil || a.me == nil {
		return &GamePacket{Cmd: ReqFrameInfo}
	}
	a.worldBound = HyperRect{Min: a.spp.Min, Max: a.spp.Max}
	a.targetlist = make(AI3AimTargetList, 0)
	a.mainobjlist = make(AI3AimTargetList, 0)
	a.spp.ApplyParts27Fn(a.prepareTarget, a.me.PosVector)

	if len(a.targetlist) == 0 {
		return &GamePacket{Cmd: ReqFrameInfo}
	}

	// for return packet
	var bulletMoveVector *Vector3D = nil
	var accvt *Vector3D = nil
	var burstCount int = 0
	var hommingTargetID []int // objid, teamid
	var superBulletMv *Vector3D = nil

	if a.ActionPoint >= GameConst.APAccel {
		softFn := func(p1, p2 *AI3AimTarget) bool {
			return p1.EvasionFactor > p2.EvasionFactor
		}
		AI3By(softFn).Sort(a.targetlist)
		for _, o := range a.targetlist {
			if o.EvasionFactor > 1 {
				accvt = a.calcEvasionVector(o.SPObj)
				a.ActionPoint -= GameConst.APAccel
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.APBullet {
		sortFn := func(p1, p2 *AI3AimTarget) bool {
			return p1.BulletAttackFactor > p2.BulletAttackFactor
		}
		AI3By(sortFn).Sort(a.targetlist)
		for _, o := range a.targetlist {
			if o.BulletAttackFactor > 1 && rand.Float64() < 0.5 {
				_, estpos, _ := a.calcAims(o.SPObj, ObjDefault.MoveLimit[GameObjBullet])

				bulletMoveVector = estpos.Sub(&a.me.PosVector).NormalizedTo(ObjDefault.MoveLimit[GameObjBullet])
				a.ActionPoint -= GameConst.APBullet
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.APSuperBullet {
		sortFn := func(p1, p2 *AI3AimTarget) bool {
			return p1.SBulletAttackFactor > p2.SBulletAttackFactor
		}
		AI3By(sortFn).Sort(a.mainobjlist)
		for _, o := range a.mainobjlist {
			if o.SBulletAttackFactor > 1 {
				//log.Printf("super %#v", o)
				_, estpos, _ := a.calcAims(o.SPObj, ObjDefault.MoveLimit[GameObjSuperBullet])
				superBulletMv = estpos.Sub(&a.me.PosVector).NormalizedTo(ObjDefault.MoveLimit[GameObjSuperBullet])
				a.ActionPoint -= GameConst.APSuperBullet
				break
			}
		}
	}
	if a.ActionPoint >= GameConst.APHommingBullet {
		sortFn := func(p1, p2 *AI3AimTarget) bool {
			return p1.HommingAttackFactor > p2.HommingAttackFactor
		}
		AI3By(sortFn).Sort(a.mainobjlist)
		for _, o := range a.mainobjlist {
			if o.HommingAttackFactor > 1 && rand.Float64() < 0.5 {
				hommingTargetID = []int{o.ID, o.TeamID}
				a.ActionPoint -= GameConst.APHommingBullet
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.APBurstShot*72 {
		burstCount = a.ActionPoint / GameConst.APBurstShot / 2
		a.ActionPoint -= GameConst.APBurstShot * burstCount
	}

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
