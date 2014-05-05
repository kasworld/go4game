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
	AimPos           *Vector3D
	AimAngle         float64
	LenToContact     float64
	NextLenToContact float64
}

func (me *SPObj) calcLens(t *SPObj) (float64, float64) {
	collen := me.CollisionRadius + t.CollisionRadius
	curlen := me.PosVector.LenTo(&t.PosVector) - collen
	nextposme := me.PosVector.Add(me.MoveVector.Idiv(60.0))
	nextpost := t.PosVector.Add(t.MoveVector.Idiv(60.0))
	nextlen := nextposme.LenTo(nextpost) - collen
	return curlen, nextlen
}

func (me *SPObj) calcAims(t *SPObj, movelimit float64) (*Vector3D, float64) {
	var estpos *Vector3D
	//durold := me.PosVector.LenTo(&t.PosVector) / movelimit // not exact
	dur := me.calcEstdur(t, movelimit)
	if dur < 0 || math.IsNaN(dur) {
		return nil, 0
	}
	//log.Printf("old %v, new %v", durold, dur)
	estpos = t.PosVector.Add(t.MoveVector.Imul(dur))
	estvt := estpos.Sub(&me.PosVector).Normalized().Imul(movelimit)
	estangle := t.MoveVector.Angle(estvt)
	return estpos, estangle
}

func (me *SPObj) calcEstdur(t *SPObj, movelimit float64) float64 {
	totargetvt := t.PosVector.Sub(&me.PosVector)
	a := t.MoveVector.Dot(&t.MoveVector) - math.Pow(movelimit, 2)
	b := 2 * t.MoveVector.Dot(totargetvt)
	c := totargetvt.Dot(totargetvt)
	p := -b / (2 * a)
	q := math.Sqrt((b*b)-4*a*c) / (2 * a)
	t1 := p - q
	t2 := p + q
	if t1 > t2 && t2 > 0 {
		return t2
	} else {
		return t1 // can - or Nan
	}
}

func (a *AIConn) calcEscapeVector(t *AimTarget) *Vector3D {
	speed := (a.me.CollisionRadius + t.SPObj.CollisionRadius) * 60
	backvt := a.me.PosVector.Sub(&t.SPObj.PosVector).Normalized().Imul(speed) // backward
	sidevt := t.AimPos.Sub(&a.me.PosVector).Normalized().Imul(speed)
	tocentervt := a.me.PosVector.Neg().Normalized().Imul(speed)

	return backvt.Add(backvt).Add(sidevt).Add(tocentervt)
}

func (a *AIConn) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			estpos, estangle := a.me.calcAims(t, 300)
			curlen, nextlen := a.me.calcLens(t)
			o := AimTarget{
				SPObj:            t,
				AimPos:           estpos,
				AimAngle:         estangle,
				LenToContact:     curlen,
				NextLenToContact: nextlen,
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
	if o.AimPos == nil || !o.AimPos.IsIn(&a.worldBound) {
		return -1.0
	}
	if o.LenToContact <= 0 || o.NextLenToContact <= 0 {
		return math.MaxFloat64
	}
	anglefactor := math.Pow(o.AimAngle/math.Pi, 2)
	typefactor := 1.0
	if o.ObjType == GameObjMain {
		typefactor = 2
	}
	lenfactor := math.Pow(o.LenToContact/o.NextLenToContact, 8)

	factor := anglefactor * typefactor * lenfactor
	// if factor > 1 {
	// 	log.Printf("fnCalcAttackFactor %v %v %v %v", anglefactor, typefactor, lenfactor, factor)
	// }
	return factor
}

// escape
func (a *AIConn) fnCalcDangerFactor(o *AimTarget) float64 {
	if !(o.ObjType == GameObjMain || o.ObjType == GameObjBullet) {
		return -1.0
	}
	if o.AimPos == nil || !o.AimPos.IsIn(&a.worldBound) {
		return -1.0
	}
	if o.LenToContact <= 0 || o.NextLenToContact <= 0 {
		return math.MaxFloat64
	}
	anglefactor := math.Pow(o.AimAngle/math.Pi, 2) // o.AimAngle/math.Pi : 0 ~ 2
	typefactor := 1.0
	if o.ObjType == GameObjMain {
		typefactor = 2
	}
	// o.LenToContact/o.NextLenToContact : 0 ~ inf
	lenfactor := math.Pow(o.LenToContact/o.NextLenToContact, 8)

	factor := anglefactor * typefactor * lenfactor
	// if factor > 1 {
	// 	log.Printf("fnCalcDangerFactor %v %v %v %v", anglefactor, typefactor, lenfactor, factor)
	// }
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
	if intertarget != nil && interval >= rand.Float64()*2 && a.ActionLimit.Bullet.Inc() {
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
	if esctarget != nil && escval >= rand.Float64()*2 && a.ActionLimit.Accel.Inc() {
		accvt = a.calcEscapeVector(esctarget)
	}
	//log.Printf("interval %v , escval %v", interval, escval)
	return &GamePacket{
		Cmd: ReqAIAct,
		ClientAct: &ClientActionPacket{
			Accel:          accvt,
			NormalBulletMv: bulletMoveVector,
		},
	}
}
