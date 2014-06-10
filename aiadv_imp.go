package go4game

import (
	//"log"
	//"fmt"
	"math"
	"math/rand"
	//"sort"
	"time"
)

var advinst2 = AIAdvFns{
	SuperFns: []AIVector3DAct{
		1: AIVector3DAct{calcSuperFactor_1, makeSuperBulletMv_1},
		4: AIVector3DAct{calcSuperFactor_4, makeSuperBulletMv_4},
	},
	BulletFns: []AIVector3DAct{
		1: AIVector3DAct{calcBulletFactor_1, makeNormalBulletMv_1},
		4: AIVector3DAct{calcBulletFactor_4, makeNormalBulletMv_4},
	},
	AccelFns: []AIVector3DAct{
		1: AIVector3DAct{calcAccelFactor_1, makeAccel_nomove},
		2: AIVector3DAct{calcAccelFactor_1, makeAccel_tohome},
		3: AIVector3DAct{calcAccelFactor_1, makeAccel_rnd},
		4: AIVector3DAct{calcAccelFactor_4, makeAccel_4},
	},
	HommingFns: []AIIDListAct{
		1: AIIDListAct{calcHommingFactor_1, makeHommingTargetID_1},
		4: AIIDListAct{calcHommingFactor_4, makeHommingTargetID_4},
	},
	BurstFns: []AIIntAct{
		1: AIIntAct{calcBurstFactor_1, makeBurstBullet_1},
		4: AIIntAct{calcBurstFactor_4, makeBurstBullet_4},
	},
}

func NewAIAdv(name string, act [ActionEnd]int) AIActor {
	a := AIAdv{
		act:      act,
		name:     name,
		AIAdvFns: advinst2,
	}
	for act := ActionAccel; act < ActionEnd; act++ {
		a.lastTargets[act] = make(map[int64]time.Time)
	}
	return &a
}

// -------------- customize begin ------------------------

// ------------- 1 -----------------------

func calcAccelFactor_1(a *AIAdv, o *AIAdvAimTarget) float64 {
	return 1
}
func makeAccel_tohome(a *AIAdv) *Vector3D {
	var vt Vector3D
	vt = a.HomePos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjMain])
	return &vt
}
func makeAccel_rnd(a *AIAdv) *Vector3D {
	var vt Vector3D
	vt = GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjMain])
	return &vt
}
func makeAccel_nomove(a *AIAdv) *Vector3D {
	vt := a.me.MoveVector.Neg()
	return &vt
}

func calcSuperFactor_1(a *AIAdv, o *AIAdvAimTarget) float64 {
	return 1
}
func makeSuperBulletMv_1(a *AIAdv) *Vector3D {
	if rand.Float64() < 0.5 {
		vt := GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjSuperBullet])
		return &vt
	}
	return nil
}

func calcBulletFactor_1(a *AIAdv, o *AIAdvAimTarget) float64 {
	return 1
}
func makeNormalBulletMv_1(a *AIAdv) *Vector3D {
	if rand.Float64() < 0.5 {
		vt := GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjBullet])
		return &vt
	}
	return nil
}

func calcHommingFactor_1(a *AIAdv, o *AIAdvAimTarget) float64 {
	return 1
}
func makeHommingTargetID_1(a *AIAdv) IDList {
	if rand.Float64() < 0.5 {
		return IDList{a.me.ID, a.me.TeamID}
	}
	return nil
}

func calcBurstFactor_1(a *AIAdv) int {
	return 40
}
func makeBurstBullet_1(a *AIAdv) int {
	return a.ActionPoint/GameConst.AP[ActionBurstBullet] - 4
}

// ---------- 4 ------------------------
func calcAccelFactor_4(a *AIAdv, o *AIAdvAimTarget) float64 {
	if !GameConst.IsInteract[a.me.ObjType][o.ObjType] {
		return -1.0
	}
	// higher is danger : 0 ~ 2
	// anglefactor := 2 - a.me.PosVector.Sub(&o.PosVector).Angle(&o.MoveVector)*2/math.Pi
	// anglefactor := 1.0

	// typefactor := [GameObjEnd]float64{
	//  GameObjMain:          2.0,
	//  GameObjBullet:        1.0,
	//  GameObjShield:        0.0,
	//  GameObjHommingBullet: 1.0,
	//  GameObjSuperBullet:   1.0,
	//  GameObjHard:          1.0,
	// }[o.ObjType]

	//speedrate := GameConst.MoveLimit[o.ObjType] / GameConst.MoveLimit[a.me.ObjType]
	// timefactor := GameConst.FramePerSec / 2 / a.frame2Contact(o) // in 0.5 sec len

	// factor := anglefactor * typefactor * lenfactor * timefactor //* speedrate
	// return factor
	lenfactor := a.calcLenRate(o.SPObj)
	return lenfactor
}
func makeAccel_4(a *AIAdv) *Vector3D {
	act := ActionAccel
	var vt Vector3D
	for i, o := range a.preparedTargets[act] {
		if i > 3 { // apply max 3 target
			break
		}
		if o.actFactor > 1 {
			vt = vt.Add(a.calcBackVector(o.SPObj, o.actFactor))
		}
	}
	if vt.Abs() < 10 {
		vt = vt.Add(a.HomePos.Sub(a.me.PosVector))
	}
	return &vt
}

func calcSuperFactor_4(a *AIAdv, o *AIAdvAimTarget) float64 {
	bulletType := GameObjSuperBullet
	if !GameConst.IsInteract[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, estangle := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
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
func makeSuperBulletMv_4(a *AIAdv) *Vector3D {
	act := ActionSuperBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[GameObjSuperBullet])
			a.lastTargets[act][o.ID] = time.Now()
			vt := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjSuperBullet])
			return &vt
		}
	}
	return nil
}

func calcBulletFactor_4(a *AIAdv, o *AIAdvAimTarget) float64 {
	bulletType := GameObjBullet
	if !GameConst.IsInteract[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, estangle := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
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
func makeNormalBulletMv_4(a *AIAdv) *Vector3D {
	act := ActionBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[GameObjBullet])
			vt := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjBullet])
			a.lastTargets[act][o.ID] = time.Now()
			return &vt
		}
	}
	return nil
}

func calcHommingFactor_4(a *AIAdv, o *AIAdvAimTarget) float64 {
	bulletType := GameObjHommingBullet
	if !GameConst.IsInteract[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, estangle := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
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
func makeHommingTargetID_4(a *AIAdv) IDList {
	act := ActionHommingBullet
	var rtn IDList
	// offencive homming
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			rtn = IDList{o.ID, o.TeamID}
			a.lastTargets[act][o.ID] = time.Now()
			break
		}
	}
	// defencive homming
	if rtn == nil {
		o := a.me
		if a.lastTargets[act][o.ID].IsZero() {
			rtn = IDList{o.ID, o.TeamID}
			a.lastTargets[act][o.ID] = time.Now()
		}
	}
	return rtn
}

func calcBurstFactor_4(a *AIAdv) int {
	return 72
}
func makeBurstBullet_4(a *AIAdv) int {
	return a.ActionPoint/GameConst.AP[ActionBurstBullet] - 4
}

// --------------------- util fns

// from gameobj moveByTimeFn_accel
func (a *AIAdv) TestMoveByAccel(m *SPObj, accelVector Vector3D) Vector3D {
	dur := 1000 / GameConst.FramePerSec
	MoveVector := m.MoveVector.Add(accelVector.Imul(dur))
	if MoveVector.Abs() > GameConst.MoveLimit[m.ObjType] {
		MoveVector = MoveVector.NormalizedTo(GameConst.MoveLimit[m.ObjType])
	}
	PosVector := m.PosVector.Add(MoveVector.Imul(dur))
	return PosVector
}

// estmate remain frame to contact( len == 0 )
func (a *AIAdv) frame2Contact(t *SPObj) float64 {
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

func (a *AIAdv) calcEvasionVector(t *SPObj) *Vector3D {
	speed := GameConst.MoveLimit[a.me.ObjType]
	backvt := a.me.PosVector.Sub(t.PosVector).NormalizedTo(speed) // backward
	tohomevt := a.HomePos.Sub(a.me.PosVector).NormalizedTo(speed) // to home pos
	rtn := backvt.Add(backvt).Add(tohomevt)
	return &rtn
}

func (a *AIAdv) calcLenRate(t *SPObj) float64 {
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

func (a *AIAdv) calcAims(t *SPObj, projectilemovelimit float64) (float64, *Vector3D, float64) {
	dur := a.me.PosVector.CalcAimAheadDur(t.PosVector, t.MoveVector, projectilemovelimit)
	if math.IsInf(dur, 1) {
		return math.Inf(1), nil, 0
	}
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estangle := t.MoveVector.Angle(estpos.Sub(a.me.PosVector))
	return dur, &estpos, estangle
}

func (a *AIAdv) calcBackVector(t *SPObj, evfactor float64) Vector3D {
	speed := GameConst.MoveLimit[a.me.ObjType]
	return a.me.PosVector.Sub(t.PosVector).NormalizedTo(evfactor * speed)
}

// -------------- customize end --------------------------
