package go4game

import (
	//"log"
	//"time"
	"math"
	"math/rand"
)

type AIConn struct {
	me          SPObj
	spp         *SpatialPartition
	ActionLimit ActStat
	targetlist  AimTargetList
	worldBound  HyperRect
}

type AimTargetList []*AimTarget

type AimTarget struct {
	*SPObj
	AimPos   *Vector3D
	AimAngle float64
	LenRate  float64
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

func (a *AIConn) calcEscapeVector(t *AimTarget) *Vector3D {
	speed := (a.me.CollisionRadius + t.SPObj.CollisionRadius) * 60
	backvt := a.me.PosVector.Sub(&t.SPObj.PosVector).NormalizedTo(speed) // backward
	sidevt := t.AimPos.Sub(&a.me.PosVector).NormalizedTo(speed)
	tocentervt := a.me.PosVector.NormalizedTo(speed / 2).Neg()

	return backvt.Add(backvt).Add(sidevt).Add(tocentervt)
}

func (a *AIConn) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			estdur, estpos, estangle := a.me.calcAims(t, 300)
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
			a.targetlist = append(a.targetlist, &o)
		}
	}
	return false
}

func (a AimTargetList) FindMax(fn func(o *AimTarget) float64) (*AimTarget, float64) {
	iv := 0.0
	var ro *AimTarget
	for _, o := range a {
		tv := fn(o)
		if tv > iv {
			iv = tv
			ro = o
		}
	}
	return ro, iv
}

// attack
func (a *AIConn) fnCalcAttackFactor(o *AimTarget) float64 {
	if !(o.ObjType == GameObjMain || o.ObjType == GameObjBullet) {
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

// escape
func (a *AIConn) fnCalcDangerFactor(o *AimTarget) float64 {
	if !(o.ObjType == GameObjMain || o.ObjType == GameObjBullet) {
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

//
func (a *AIConn) makeAIAction() *GamePacket {
	if a.spp == nil {
		return &GamePacket{Cmd: ReqAIAct}
	}
	a.worldBound = HyperRect{Min: a.spp.Min, Max: a.spp.Max}
	a.targetlist = make(AimTargetList, 0)
	a.spp.ApplyParts27Fn(a.prepareTarget, a.me.PosVector)

	var bulletMoveVector *Vector3D
	var accvt *Vector3D

	intertarget, interval := a.targetlist.FindMax(a.fnCalcAttackFactor)
	if intertarget != nil && interval >= (rand.Float64()+1) && a.ActionLimit.Bullet.Inc() {
		//log.Printf("attack %v", interval)
		var aimpos *Vector3D
		if intertarget.ObjType != GameObjMain {
			aimpos = intertarget.AimPos
		} else {
			// add random ness to target pos
			aimpos = intertarget.AimPos
			//aimpos = intertarget.AimPos.Add(RandVector3D(0.0, intertarget.CollisionRadius*4))
		}

		bulletMoveVector = aimpos.Sub(&a.me.PosVector).NormalizedTo(300.0)
	}

	esctarget, escval := a.targetlist.FindMax(a.fnCalcDangerFactor)
	if esctarget != nil && escval >= (rand.Float64()+1) && a.ActionLimit.Accel.Inc() {
		//log.Printf("escval %v", escval)
		accvt = a.calcEscapeVector(esctarget)
	}

	return &GamePacket{
		Cmd: ReqAIAct,
		ClientAct: &ClientActionPacket{
			Accel:          accvt,
			NormalBulletMv: bulletMoveVector,
		},
	}
}
