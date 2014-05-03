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
	targetlist  SPObjList
}

func (a *AIConn) appendTarget(s SPObjList) bool {
	for _, o := range s {
		if a.me.TeamID != o.TeamID {
			a.targetlist = append(a.targetlist, o)
		}
	}
	return false
}

// estimate target pos by target speed
func (me *SPObj) calcAimMvVector(t *SPObj, movelimit float64) *Vector3D {
	dur := me.PosVector.LenTo(&t.PosVector) / movelimit
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	return estpos.Sub(&me.PosVector).Normalized().Imul(movelimit)
}

func (me *SPObj) calcDangerByVt(t *SPObj) float64 {
	collen := me.CollisionRadius - t.CollisionRadius
	curlen := me.PosVector.Sub(&t.PosVector).Abs() - collen
	nextlen := me.PosVector.Add(me.MoveVector.Idiv(60.0)).Sub(
		t.PosVector.Add(t.MoveVector.Idiv(60.0))).Abs() - collen
	if nextlen <= 0 || curlen <= 0 {
		return math.MaxFloat64
	}
	return curlen / nextlen
}

func (me *SPObj) calcDangerByLen(t *SPObj) float64 {
	return me.PosVector.LenTo(&t.PosVector) / (me.CollisionRadius + t.CollisionRadius)
}

func (me *SPObj) calcEscapeVector(t *SPObj) *Vector3D {
	speed := (me.CollisionRadius + t.CollisionRadius) * 60
	vt := me.PosVector.Sub(&t.PosVector).Normalized().Imul(speed)
	vt = vt.Sub(me.PosVector.Normalized().Imul(speed / 2))
	//vt = vt.Add(RandVector3D(speed/2, speed))
	return vt
}

func (a *AIConn) selectDangerTarget() (*SPObj, *SPObj) {
	var odt *SPObj
	var odl *SPObj
	maxdt := 0.0
	mindl := a.spp.Max.Abs()
	for _, o := range a.targetlist {
		if o.ObjType == GameObjMain || o.ObjType == GameObjBullet {
			dt := a.me.calcDangerByVt(o)
			if dt > maxdt {
				maxdt = dt
				odt = o
			}
			dl := a.me.calcDangerByLen(o)
			if dl < mindl {
				mindl = dl
				odl = o
			}
		}
	}
	return odt, odl
}

func (a *AIConn) makeAIAction() *GamePacket {
	if a.spp == nil {
		return &GamePacket{Cmd: ReqAIAct}
	}
	a.targetlist = make(SPObjList, 0)
	a.spp.ApplyParts27Fn(a.appendTarget, a.me.PosVector, a.spp.MaxObjectRadius)

	var bulletMoveVector *Vector3D
	var accvt *Vector3D
	dtvt, dtl := a.selectDangerTarget()
	if dtvt != nil {
		if a.me.calcDangerByVt(dtvt) > 1 && rand.Float64() < 0.5 && a.ActionLimit.Bullet.Inc() {
			bulletMoveVector = a.me.calcAimMvVector(dtvt, 300)
		}
		if a.me.calcDangerByLen(dtl) < 10 && rand.Float64() < 0.5 && a.ActionLimit.Accel.Inc() {
			accvt = a.me.calcEscapeVector(dtvt)
		}

		//log.Printf("acc escape %v ", accvt)
	}

	return &GamePacket{
		Cmd: ReqAIAct,
		ClientAct: &ClientActionPacket{
			Accel:          accvt,
			NormalBulletMv: bulletMoveVector,
		},
	}
}
