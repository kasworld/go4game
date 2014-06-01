package go4game

import (
	//"log"
	//"time"
	"math"
	"math/rand"
	"sort"
)

type AI2 struct {
	me          *SPObj
	spp         *SpatialPartition
	ActionPoint int
	Score       int
	HomePos     Vector3D

	targetlist  AimTargetList
	mainobjlist AimTargetList
}

func NewAI2() AIActor {
	return &AI2{}
}

func (a *AI2) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			estdur, estpos, estangle := a.me.calcAims(t, GameConst.MoveLimit[t.ObjType])
			if math.IsInf(estdur, 1) || !estpos.IsIn(&a.spp.WorldCube) {
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
func (a *AI2) MakeAction(packet *RspGamePacket) *ReqGamePacket {
	a.spp = packet.Spp
	a.me = packet.TeamInfo.SPObj
	a.ActionPoint = packet.TeamInfo.ActionPoint
	a.Score = packet.TeamInfo.Score
	a.HomePos = packet.TeamInfo.HomePos

	if a.spp == nil || a.me == nil {
		return &ReqGamePacket{Cmd: ReqFrameInfo}
	}
	a.targetlist = make(AimTargetList, 0)
	a.mainobjlist = make(AimTargetList, 0)
	a.spp.ApplyParts27Fn(a.prepareTarget, a.me.PosVector)

	if len(a.targetlist) == 0 {
		return &ReqGamePacket{Cmd: ReqFrameInfo}
	}

	// for return packet
	var bulletMoveVector *Vector3D = nil
	var accvt *Vector3D = nil
	var burstCount int = 0
	var hommingTargetID IDList // objid, teamid
	var superBulletMv *Vector3D = nil

	if a.ActionPoint >= GameConst.AP[ActionSuperBullet] {
		attackFn := func(p1, p2 *AimTarget) bool {
			return p1.AttackFactor > p2.AttackFactor
		}
		By(attackFn).Sort(a.mainobjlist)
		for _, o := range a.mainobjlist {
			if o.AttackFactor > 1 && rand.Float64() < 0.5 {
				t := o.AimPos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjSuperBullet])
				superBulletMv = &t
				a.ActionPoint -= GameConst.AP[ActionSuperBullet]
				break
			}
		}
	}
	if a.ActionPoint >= GameConst.AP[ActionHommingBullet] {
		attackFn := func(p1, p2 *AimTarget) bool {
			return p1.AttackFactor > p2.AttackFactor
		}
		By(attackFn).Sort(a.mainobjlist)
		for _, o := range a.mainobjlist {
			if o.AttackFactor > 1 && rand.Float64() < 0.5 {
				hommingTargetID = IDList{o.ID, o.TeamID}
				a.ActionPoint -= GameConst.AP[ActionHommingBullet]
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.AP[ActionAccel] {
		evasionFn := func(p1, p2 *AimTarget) bool {
			return p1.EvasionFactor > p2.EvasionFactor
		}
		By(evasionFn).Sort(a.targetlist)
		for _, o := range a.targetlist {
			if o.EvasionFactor > 1 && rand.Float64() < 0.9 {
				accvt = a.calcEvasionVector(o)
				a.ActionPoint -= GameConst.AP[ActionAccel]
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.AP[ActionBullet] {
		attackFn := func(p1, p2 *AimTarget) bool {
			return p1.AttackFactor > p2.AttackFactor
		}
		By(attackFn).Sort(a.targetlist)
		for _, o := range a.targetlist {
			if o.AttackFactor > 1 && rand.Float64() < 0.5 {
				tmpbulletMoveVector := o.AimPos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjBullet])
				bulletMoveVector = &tmpbulletMoveVector
				a.ActionPoint -= GameConst.AP[ActionBullet]
				break
			}
		}
	}

	if a.ActionPoint >= GameConst.AP[ActionBurstBullet]*40 {
		burstCount = a.ActionPoint/GameConst.AP[ActionBurstBullet] - 4
		a.ActionPoint -= GameConst.AP[ActionBurstBullet] * burstCount
	}

	return &ReqGamePacket{
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
	speed := math.Sqrt(GameConst.ObjSqd[a.me.ObjType][t.ObjType]) * GameConst.FramePerSec
	backvt := a.me.PosVector.Sub(t.SPObj.PosVector).NormalizedTo(speed) // backward
	sidevt := t.AimPos.Sub(a.me.PosVector).NormalizedTo(speed)
	tohomevt := a.HomePos.Sub(a.me.PosVector).NormalizedTo(speed) // to home pos
	rtn := backvt.Add(backvt).Add(sidevt).Add(tohomevt)
	return &rtn
}

// attack
func (a *AI2) CalcBulletAttackFactor(o *AimTarget) float64 {
	// is obj attacked by bullet?
	if !GameConst.IsInteract[o.ObjType][GameObjBullet] {
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
	if !GameConst.IsInteract[GameObjMain][o.ObjType] {
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
	collen := math.Sqrt(GameConst.ObjSqd[me.ObjType][t.ObjType])
	curlen := me.PosVector.LenTo(t.PosVector) - collen
	nextposme := me.PosVector.Add(me.MoveVector.Idiv(GameConst.FramePerSec))
	nextpost := t.PosVector.Add(t.MoveVector.Idiv(GameConst.FramePerSec))
	nextlen := nextposme.LenTo(nextpost) - collen
	if curlen <= 0 || nextlen <= 0 {
		return math.Inf(1)
	} else {
		return curlen / nextlen
	}
}

//
func (me *SPObj) calcAims(t *SPObj, movelimit float64) (float64, *Vector3D, float64) {
	dur := me.PosVector.CalcAimAheadDur(t.PosVector, t.MoveVector, movelimit)
	if math.IsInf(dur, 1) {
		return math.Inf(1), nil, 0
	}
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estangle := t.MoveVector.Angle(estpos.Sub(me.PosVector))
	return dur, &estpos, estangle
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
