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
	AimVt            *Vector3D
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

func (me *SPObj) calcAims(t *SPObj, movelimit float64) (*Vector3D, *Vector3D, float64) {
	dur := me.PosVector.LenTo(&t.PosVector) / movelimit
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estvt := estpos.Sub(&me.PosVector).Normalized().Imul(movelimit)
	estangle := t.MoveVector.Angle(estvt)
	return estpos, estvt, estangle
}

func (v *Vector3D) sideWalk(dst *Vector3D) *Vector3D {
	tovt := dst.Sub(v)
	return tovt.Cross(dst)
}

func (me *SPObj) calcEscapeVector(t *SPObj) *Vector3D {
	speed := (me.CollisionRadius + t.CollisionRadius) * 60
	vt := me.PosVector.Sub(&t.PosVector).Normalized().Imul(speed)
	vt = vt.Add(me.PosVector.sideWalk(&t.PosVector).Normalized().Imul(speed / 2))
	vt = vt.Sub(me.PosVector.Normalized().Imul(speed / 2)) // add to center
	return vt
}

func (a *AIConn) prepareTarget(s SPObjList) bool {
	for _, t := range s {
		if a.me.TeamID != t.TeamID {
			estpos, estvt, estangle := a.me.calcAims(t, 300)
			curlen, nextlen := a.me.calcLens(t)
			o := AimTarget{
				SPObj:            t,
				AimPos:           estpos,
				AimVt:            estvt,
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

func (a *AIConn) fnCalcValueIntercept(o *AimTarget) float64 {
	if !(o.ObjType == GameObjMain || o.ObjType == GameObjBullet) {
		return -1.0
	}
	if !o.AimPos.IsIn(&a.worldBound) {
		return -1.0
	}
	if o.LenToContact <= 0 || o.NextLenToContact <= 0 {
		return math.MaxFloat64
	}
	anglefactor := math.Pow(o.AimAngle/math.Pi, 2)
	typefactor := 1.0
	if o.ObjType == GameObjMain {
		typefactor = 1.4
	}
	lenfactor := math.Pow(o.LenToContact/o.NextLenToContact, 4)

	factor := anglefactor * typefactor * lenfactor
	// if factor > 1 {
	// 	log.Printf("fnCalcValueIntercept %v %v %v %v", anglefactor, typefactor, lenfactor, factor)
	// }
	return factor
}

func (a *AIConn) fnCalcValueEscape(o *AimTarget) float64 {
	if !(o.ObjType == GameObjMain || o.ObjType == GameObjBullet) {
		return -1.0
	}
	if !o.AimPos.IsIn(&a.worldBound) {
		return -1.0
	}
	if o.LenToContact <= 0 || o.NextLenToContact <= 0 {
		return math.MaxFloat64
	}
	anglefactor := math.Pow(o.AimAngle/math.Pi, 4)
	typefactor := 1.0
	if o.ObjType == GameObjMain {
		typefactor = 1.1
	}
	lenfactor := math.Pow(o.LenToContact/o.NextLenToContact, 8)

	factor := anglefactor * typefactor * lenfactor
	// if factor > 1 {
	// 	log.Printf("fnCalcValueEscape %v %v %v %v", anglefactor, typefactor, lenfactor, factor)
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

	intertarget, interval := a.targetlist.FindMax(a.fnCalcValueIntercept)
	if intertarget != nil && interval >= rand.Float64()*2 && a.ActionLimit.Bullet.Inc() {
		bulletMoveVector = intertarget.AimVt
	}

	esctarget, escval := a.targetlist.FindMax(a.fnCalcValueEscape)
	if esctarget != nil && escval >= rand.Float64()*2 && a.ActionLimit.Accel.Inc() {
		accvt = a.me.calcEscapeVector(esctarget.SPObj)
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
