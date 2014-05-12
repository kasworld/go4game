package go4game

import (
	//"log"
	//"time"
	"math"
	"math/rand"
	"sort"
)

// AINothing ----------------------------------------------------------------
type AINothing struct {
}

func (a *AINothing) MakeAction(packet *GamePacket) *GamePacket {
	var bulletMoveVector *Vector3D = nil
	var accvt *Vector3D = nil
	var burstCount int = 0
	var hommingTargetID []int // objid, teamid
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
	worldBound  HyperRect
	ActionPoint int
	Score       int
}

func (a *AIRandom) MakeAction(packet *GamePacket) *GamePacket {
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	if a.spp == nil {
		return &GamePacket{Cmd: ReqFrameInfo}
	}
	a.worldBound = HyperRect{Min: a.spp.Min, Max: a.spp.Max}

	var bulletMoveVector *Vector3D = nil
	var accvt *Vector3D = nil
	var burstCount int = 0
	var hommingTargetID []int // objid, teamid
	var superBulletMv *Vector3D = nil

	if a.ActionPoint >= GameConst.APAccel {
		accvt = RandVector(a.spp.Min, a.spp.Max)
		a.ActionPoint -= GameConst.APAccel
	}

	if a.ActionPoint >= GameConst.APBullet {
		bulletMoveVector = RandVector(a.spp.Min, a.spp.Max)
		a.ActionPoint -= GameConst.APBullet
	}

	if a.ActionPoint >= GameConst.APSuperBullet {
		superBulletMv = RandVector(a.spp.Min, a.spp.Max)
		a.ActionPoint -= GameConst.APSuperBullet
	}
	// if a.ActionPoint >= GameConst.APHommingBullet {
	// 	hommingTargetID = []int{o.ID, o.TeamID}
	// 	a.ActionPoint -= GameConst.APHommingBullet
	// }

	if a.ActionPoint >= GameConst.APBurstShot*40 {
		burstCount = a.ActionPoint/GameConst.APBurstShot - 4
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

// AI2 ----------------------------------------------------------------
type AI2 struct {
	me          *SPObj
	spp         *SpatialPartition
	worldBound  HyperRect
	ActionPoint int
	Score       int

	targetlist  AimTargetList
	mainobjlist AimTargetList
}

func (a *AI2) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			estdur, estpos, estangle := a.me.calcAims(t, ObjDefault.MoveLimit[t.ObjType])
			if math.IsInf(estdur, 1) || !estpos.IsIn(&a.worldBound) {
				estpos = nil
			}
			lenRate := a.me.calcLenRate(t)
			o := AimTarget{
				SPObj:    t,
				AimPos:   estpos,
				AimAngle: estangle,
				LenRate:  lenRate,
			}
			o.AttackFactor = a.CalcBulletAttackFactor(&o)
			o.EvasionFactor = a.CalcEvasionFactor(&o)
			a.targetlist = append(a.targetlist, &o)

			if t.ObjType == GameObjMain {
				a.mainobjlist = append(a.mainobjlist, &o)
			}
		}
	}
	return false
}
func (a *AI2) MakeAction(packet *GamePacket) *GamePacket {
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score

	if a.spp == nil || a.me == nil {
		return &GamePacket{Cmd: ReqFrameInfo}
	}
	a.worldBound = HyperRect{Min: a.spp.Min, Max: a.spp.Max}
	a.targetlist = make(AimTargetList, 0)
	a.mainobjlist = make(AimTargetList, 0)
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
		evasionFn := func(p1, p2 *AimTarget) bool {
			return p1.EvasionFactor > p2.EvasionFactor
		}
		By(evasionFn).Sort(a.targetlist)
		for _, o := range a.targetlist {
			if o.EvasionFactor > 1 && rand.Float64() < 0.9 {
				accvt = a.calcEvasionVector(o)
				a.ActionPoint -= GameConst.APAccel
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.APBullet {
		attackFn := func(p1, p2 *AimTarget) bool {
			return p1.AttackFactor > p2.AttackFactor
		}
		By(attackFn).Sort(a.targetlist)
		for _, o := range a.targetlist {
			if o.AttackFactor > 1 && rand.Float64() < 0.5 {
				bulletMoveVector = o.AimPos.Sub(&a.me.PosVector).NormalizedTo(ObjDefault.MoveLimit[GameObjBullet])
				a.ActionPoint -= GameConst.APBullet
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.APSuperBullet {
		attackFn := func(p1, p2 *AimTarget) bool {
			return p1.AttackFactor > p2.AttackFactor
		}
		By(attackFn).Sort(a.mainobjlist)
		for _, o := range a.mainobjlist {
			if o.AttackFactor > 1 && rand.Float64() < 0.5 {
				superBulletMv = o.AimPos.Sub(&a.me.PosVector).NormalizedTo(ObjDefault.MoveLimit[GameObjSuperBullet])
				a.ActionPoint -= GameConst.APSuperBullet
				break
			}
		}
	}
	if a.ActionPoint >= GameConst.APHommingBullet {
		attackFn := func(p1, p2 *AimTarget) bool {
			return p1.AttackFactor > p2.AttackFactor
		}
		By(attackFn).Sort(a.mainobjlist)
		for _, o := range a.mainobjlist {
			if o.AttackFactor > 1 && rand.Float64() < 0.5 {
				hommingTargetID = []int{o.ID, o.TeamID}
				a.ActionPoint -= GameConst.APHommingBullet
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.APBurstShot*40 {
		burstCount = a.ActionPoint/GameConst.APBurstShot - 4
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
func (a *AI2) calcEvasionVector(t *AimTarget) *Vector3D {
	speed := (a.me.CollisionRadius + t.SPObj.CollisionRadius) * 60
	backvt := a.me.PosVector.Sub(&t.SPObj.PosVector).NormalizedTo(speed) // backward
	sidevt := t.AimPos.Sub(&a.me.PosVector).NormalizedTo(speed)
	tocentervt := a.me.PosVector.NormalizedTo(speed / 2).Neg()
	return backvt.Add(backvt).Add(sidevt).Add(tocentervt)
}

// attack
func (a *AI2) CalcBulletAttackFactor(o *AimTarget) float64 {
	// is obj attacked by bullet?
	if !InteractionMap[o.ObjType][GameObjBullet] {
		return -1.0
	}
	if o.AimPos == nil {
		return -1.0
	}
	anglefactor := math.Pow(o.AimAngle/math.Pi, 2)
	typefactor := 1.0
	if o.ObjType == GameObjMain {
		typefactor = 3
	}
	lenfactor := math.Pow(o.LenRate, 8)

	factor := anglefactor * typefactor * lenfactor
	return factor
}

// evasion
func (a *AI2) CalcEvasionFactor(o *AimTarget) float64 {
	// can obj attact me?
	if !InteractionMap[GameObjMain][o.ObjType] {
		return -1.0
	}
	if o.AimPos == nil {
		return -1.0
	}
	anglefactor := math.Pow(o.AimAngle/math.Pi, 2)
	typefactor := 1.0
	if o.ObjType == GameObjMain {
		typefactor = 1.5
	}
	lenfactor := math.Pow(o.LenRate, 8)

	factor := anglefactor * typefactor * lenfactor
	return factor
}

// ai utils -----------------------------------------------------------

type AimTargetList []*AimTarget

type AimTarget struct {
	*SPObj
	AimPos        *Vector3D
	AimAngle      float64
	LenRate       float64
	AttackFactor  float64
	EvasionFactor float64
}

// how fast collision occur
// < 1 safe , > 1 danger
func (me *SPObj) calcLenRate(t *SPObj) float64 {
	collen := me.CollisionRadius + t.CollisionRadius
	curlen := me.PosVector.LenTo(&t.PosVector) - collen
	nextposme := me.PosVector.Add(me.MoveVector.Idiv(60.0))
	nextpost := t.PosVector.Add(t.MoveVector.Idiv(60.0))
	nextlen := nextposme.LenTo(nextpost) - collen
	if curlen <= 0 || nextlen <= 0 {
		return math.Inf(1)
	} else {
		return curlen / nextlen
	}
}

//
func (me *SPObj) calcAims(t *SPObj, movelimit float64) (float64, *Vector3D, float64) {
	dur := me.PosVector.CalcAimAheadDur(&t.PosVector, &t.MoveVector, movelimit)
	if math.IsInf(dur, 1) {
		return math.Inf(1), nil, 0
	}
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estangle := t.MoveVector.Angle(estpos.Sub(&me.PosVector))
	return dur, estpos, estangle
}

// By is the type of a "less" function that defines the ordering of its AimTarget arguments.
type By func(p1, p2 *AimTarget) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(aimtargets AimTargetList) {
	ps := &aimtargetSorter{
		aimtargets: aimtargets,
		by:         by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// aimtargetSorter joins a By function and a slice of AimTargets to be sorted.
type aimtargetSorter struct {
	aimtargets AimTargetList
	by         func(p1, p2 *AimTarget) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *aimtargetSorter) Len() int {
	return len(s.aimtargets)
}

// Swap is part of sort.Interface.
func (s *aimtargetSorter) Swap(i, j int) {
	s.aimtargets[i], s.aimtargets[j] = s.aimtargets[j], s.aimtargets[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *aimtargetSorter) Less(i, j int) bool {
	return s.by(s.aimtargets[i], s.aimtargets[j])
}
