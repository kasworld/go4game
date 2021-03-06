package shootbase

import (
	"log"
	//"fmt"
	"math"
	"math/rand"
	//"sort"
	"github.com/kasworld/go4game"
	"strconv"
	"strings"
	"time"
)

var advinst2 = AIAdvFns{
	AccelFns: []AIVector3DAct{
		1: AIVector3DAct{nil, makeAccel_nomove},
		2: AIVector3DAct{nil, makeAccel_tohome},
		3: AIVector3DAct{nil, makeAccel_rnd},
		4: AIVector3DAct{calcAccelFactor_wideSpace, makeAccel_wideSpace},
		5: AIVector3DAct{nil, makeAccel_nearSight},
	},
	BulletFns: []AIVector3DAct{
		1: AIVector3DAct{nil, makeNormalBulletMv_1},
		3: AIVector3DAct{calcBulletFactor_4, makeNormalBulletMv_4_adj},
		4: AIVector3DAct{calcBulletFactor_4, makeNormalBulletMv_4},
		5: AIVector3DAct{calcBulletFactor_5, makeNormalBulletMv_5},
	},
	SuperFns: []AIVector3DAct{
		1: AIVector3DAct{nil, makeSuperBulletMv_1},
		3: AIVector3DAct{calcSuperFactor_4, makeSuperBulletMv_4_adj},
		4: AIVector3DAct{calcSuperFactor_4, makeSuperBulletMv_4},
		5: AIVector3DAct{calcSuperFactor_5, makeSuperBulletMv_5},
	},
	HommingFns: []AIIDListAct{
		1: AIIDListAct{nil, makeHommingTargetID_fullshield},
		2: AIIDListAct{calcHommingFactor_4, makeHommingTargetID_halfshield},
	},
	BurstFns: []AIIntAct{
		1: AIIntAct{calcBurstFactor_30, makeBurstBullet_sub4},
		2: AIIntAct{calcBurstFactor_52, makeBurstBullet_div2},
	},
}

func AIstr2AIActor(aistr string) AIActor {
	sname := strings.Split(aistr, "-")
	act := [5]int{}

	for i := 0; i < 5; i++ {
		v, err := strconv.Atoi(sname[i+1])
		if err != nil {
			log.Printf("unknown AI %v", aistr)
			return nil
		}
		act[i] = v
	}
	return NewAIAdv(sname[0], act)
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

// ---------- burst ------
func calcBurstFactor_30(a *AIAdv) int {
	return 30
}
func calcBurstFactor_52(a *AIAdv) int {
	return 52
}
func makeBurstBullet_sub4(a *AIAdv) int {
	return a.ActionPoint/GameConst.AP[ActionBurstBullet] - 4
}
func makeBurstBullet_div2(a *AIAdv) int {
	return a.ActionPoint / GameConst.AP[ActionBurstBullet] / 2
}

// ------------ homming -----------
func makeHommingTargetID_fullshield(a *AIAdv) go4game.IDList {
	if rand.Float64() < 0.5 {
		return go4game.IDList{a.me.ID, a.me.TeamID}
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
func makeHommingTargetID_halfshield(a *AIAdv) go4game.IDList {
	act := ActionHommingBullet
	var rtn go4game.IDList
	// offencive homming
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			rtn = go4game.IDList{o.ID, o.TeamID}
			a.lastTargets[act][o.ID] = time.Now()
			break
		}
	}
	// defencive homming
	if rtn == nil {
		o := a.me
		if a.lastTargets[act][o.ID].IsZero() {
			rtn = go4game.IDList{o.ID, o.TeamID}
			a.lastTargets[act][o.ID] = time.Now()
		}
	}
	return rtn
}

// ------------- accel -----------------------

func makeAccel_tohome(a *AIAdv) *go4game.Vector3D {
	var vt go4game.Vector3D
	vt = a.HomePos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjMain])
	return &vt
}
func makeAccel_rnd(a *AIAdv) *go4game.Vector3D {
	var vt go4game.Vector3D
	vt = GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjMain])
	return &vt
}
func makeAccel_nomove(a *AIAdv) *go4game.Vector3D {
	vt := a.me.MoveVector.Neg().Imul(GameConst.FramePerSec)
	return &vt
}

func calcAccelFactor_wideSpace(a *AIAdv, o *AIAdvAimTarget) float64 {
	if !GameConst.IsInteract[a.me.ObjType][o.ObjType] {
		return -1.0
	}
	// higher is danger : 0 ~ 2
	// anglefactor := 2 - a.me.PosVector.Sub(&o.PosVector).Angle(&o.MoveVector)*2/math.Pi
	anglefactor := 1.0

	lenfactor := a.calcLenRate(o.SPObj)

	typefactor := [GameObjEnd]float64{
		GameObjMain:          2.0,
		GameObjBullet:        1.0,
		GameObjShield:        0.0,
		GameObjHommingBullet: 1.0,
		GameObjSuperBullet:   1.0,
		GameObjHard:          1.0,
	}[o.ObjType]

	//speedrate := GameConst.MoveLimit[o.ObjType] / GameConst.MoveLimit[a.me.ObjType]
	timefactor := GameConst.FramePerSec / 2 / a.frame2Contact(o.SPObj) // in 0.5 sec len

	factor := anglefactor * typefactor * lenfactor * timefactor //* speedrate
	return factor
	// return lenfactor
}

func makeAccel_wideSpace(a *AIAdv) *go4game.Vector3D {
	act := ActionAccel
	var vt go4game.Vector3D
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

func makeAccel_nearSight(a *AIAdv) *go4game.Vector3D {
	act := ActionAccel
	toHomeVt := a.HomePos.Sub(a.me.PosVector).Imul(GameConst.FramePerSec)
	vt := toHomeVt
	for _, o := range a.preparedTargets[act] {
		if a.me.calcLenFactor(o.SPObj) >= 1 {
			spdFactor := 1.0
			if GameConst.MoveLimit[a.me.ObjType] < GameConst.MoveLimit[o.ObjType] {
				spdFactor += GameConst.MoveLimit[o.ObjType] / GameConst.MoveLimit[a.me.ObjType]
			}
			dst2srcVt := vt.Project(o.PosVector.Sub(a.me.PosVector)).Neg()
			vtfatordanger := dst2srcVt.Imul(spdFactor)
			vt = vt.Add(vtfatordanger)
		}
	}
	revMv := a.me.MoveVector.Neg().Imul(GameConst.FramePerSec)
	vt = vt.Add(revMv)
	return &vt
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

func (a *AIAdv) calcBackVector(t *SPObj, evfactor float64) go4game.Vector3D {
	speed := GameConst.MoveLimit[a.me.ObjType]
	return a.me.PosVector.Sub(t.PosVector).NormalizedTo(evfactor * speed)
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

func (m *SPObj) colilen(t *SPObj) (float64, float64) {
	collen := GameConst.Radius[m.ObjType] + GameConst.Radius[t.ObjType]
	curlen := m.PosVector.LenTo(t.PosVector) - collen
	return curlen, collen
}

func (s *SPObj) calcLenFactor(o *SPObj) float64 {
	// safe <- -1, 0, 1, 2, 3 -> danger
	if !GameConst.IsInteract[s.ObjType][o.ObjType] {
		return -1.0
	}
	spdFactor := 1.0
	if GameConst.MoveLimit[s.ObjType] < GameConst.MoveLimit[o.ObjType] {
		spdFactor += GameConst.MoveLimit[o.ObjType] / GameConst.MoveLimit[s.ObjType]
	}

	curlen, collen := s.colilen(o)
	nextlen, _ := s.TestMoveByAccel(go4game.V3DZero).colilen(o.TestMoveByAccel(go4game.V3DZero))
	lenrange := (collen / 2) * spdFactor
	var lenfactor float64
	if curlen <= lenrange {
		lenfactor += 1
	}
	if nextlen <= lenrange {
		lenfactor += 2
	}
	return lenfactor
}

// from gameobj moveByTimeFn_accel
func (m SPObj) TestMoveByAccel(accelVector go4game.Vector3D) *SPObj {
	dur := 1.0 / GameConst.FramePerSec
	m.MoveVector = m.MoveVector.Add(accelVector.Imul(dur))
	if m.MoveVector.Abs() > GameConst.MoveLimit[m.ObjType] {
		m.MoveVector = m.MoveVector.NormalizedTo(GameConst.MoveLimit[m.ObjType])
	}
	m.PosVector = m.PosVector.Add(m.MoveVector.Imul(dur))
	return &m
}

// ------------ fire ------------------

func makeSuperBulletMv_1(a *AIAdv) *go4game.Vector3D {
	if rand.Float64() < 0.5 {
		vt := GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjSuperBullet])
		return &vt
	}
	return nil
}
func makeNormalBulletMv_1(a *AIAdv) *go4game.Vector3D {
	if rand.Float64() < 0.5 {
		vt := GameConst.WorldCube.RandVector().Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[GameObjBullet])
		return &vt
	}
	return nil
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
func makeSuperBulletMv_4(a *AIAdv) *go4game.Vector3D {
	act := ActionSuperBullet
	bulletType := GameObjSuperBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
			a.lastTargets[act][o.ID] = time.Now()
			vt := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[bulletType])
			return &vt
		}
	}
	return nil
}
func makeSuperBulletMv_4_adj(a *AIAdv) *go4game.Vector3D {
	act := ActionSuperBullet
	bulletType := GameObjSuperBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
			vt := a.AimAdjedIntoCube(estpos, o.SPObj, bulletType)
			if vt == nil {
				continue
			}
			a.lastTargets[act][o.ID] = time.Now()
			return vt
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
func makeNormalBulletMv_4(a *AIAdv) *go4game.Vector3D {
	act := ActionBullet
	bulletType := GameObjBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
			vt := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[bulletType])
			a.lastTargets[act][o.ID] = time.Now()
			return &vt
		}
	}
	return nil
}
func makeNormalBulletMv_4_adj(a *AIAdv) *go4game.Vector3D {
	act := ActionBullet
	bulletType := GameObjBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 && a.lastTargets[act][o.ID].IsZero() {
			_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
			vt := a.AimAdjedIntoCube(estpos, o.SPObj, bulletType)
			if vt == nil {
				continue
			}
			a.lastTargets[act][o.ID] = time.Now()
			return vt
		}
	}
	return nil
}

func calcBulletFactor_5(a *AIAdv, o *AIAdvAimTarget) float64 {
	bulletType := GameObjBullet
	if !GameConst.IsInteract[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
	if estpos == nil || (!estpos.IsIn(GameConst.WorldCube2) && o.ObjType != GameObjMain) { // cannot contact
		return -1.0
	}

	mvangle := a.me.PosVector.Sub(o.PosVector).Angle(o.MoveVector)
	anglefactor := mvangle / math.Pi * 180
	if anglefactor > 30 && o.ObjType != GameObjMain {
		return -1.0
	}
	//log.Printf("af %v %v", estangle, anglefactor)

	//collen := math.Sqrt(GameConst.ObjSqd[a.me.ObjType][o.ObjType])
	curlen := a.me.PosVector.LenTo(o.PosVector)
	lenfactor := GameConst.WorldDiag / 2 / curlen
	//log.Printf("lf %v ", lenfactor)

	factor := lenfactor
	return factor
}
func makeNormalBulletMv_5(a *AIAdv) *go4game.Vector3D {
	act := ActionBullet
	bulletType := GameObjBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 {
			if a.lastTargets[act][o.ID].IsZero() {
				_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])

				vt := a.AimAdjedIntoCube(estpos, o.SPObj, bulletType)
				if vt == nil {
					continue
				}
				a.lastTargets[act][o.ID] = time.Now()
				return vt
			}
		}
	}
	return nil
}

func calcSuperFactor_5(a *AIAdv, o *AIAdvAimTarget) float64 {
	bulletType := GameObjSuperBullet
	if !GameConst.IsInteract[o.ObjType][bulletType] {
		return -1.0
	}
	_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
	if estpos == nil || (!estpos.IsIn(GameConst.WorldCube2) && o.ObjType != GameObjMain) { // cannot contact
		return -1.0
	}

	mvangle := a.me.PosVector.Sub(o.PosVector).Angle(o.MoveVector)
	anglefactor := mvangle / math.Pi * 180
	if anglefactor > 30 && o.ObjType != GameObjMain {
		return -1.0
	}
	//log.Printf("af %v %v", estangle, anglefactor)

	typefactor := [GameObjEnd]float64{
		GameObjMain:          1.,
		GameObjBullet:        0.5,
		GameObjShield:        0,
		GameObjHommingBullet: 1.,
		GameObjSuperBullet:   1.,
	}[o.ObjType]

	//collen := math.Sqrt(GameConst.ObjSqd[a.me.ObjType][o.ObjType])
	curlen := a.me.PosVector.LenTo(o.PosVector)
	lenfactor := GameConst.WorldDiag / 2 / curlen
	//log.Printf("lf %v ", lenfactor)

	factor := lenfactor * typefactor
	return factor
}
func makeSuperBulletMv_5(a *AIAdv) *go4game.Vector3D {
	act := ActionSuperBullet
	bulletType := GameObjSuperBullet
	for _, o := range a.preparedTargets[act] {
		if o.actFactor > 1 {
			if a.lastTargets[act][o.ID].IsZero() {
				_, estpos, _ := a.calcAims(o.SPObj, GameConst.MoveLimit[bulletType])
				vt := a.AimAdjedIntoCube(estpos, o.SPObj, bulletType)
				if vt == nil {
					continue
				}
				a.lastTargets[act][o.ID] = time.Now()
				return vt
			}
		}
	}
	return nil
}

func (a *AIAdv) AimAdjedIntoCube(estpos *go4game.Vector3D, o *SPObj, bulletType GameObjectType) *go4game.Vector3D {
	if !estpos.IsIn(GameConst.WorldCube2) && o.ObjType != GameObjMain {
		return nil
	}
	lenori := a.me.PosVector.LenTo(*estpos)
	if o.ObjType == GameObjMain {
		changed := estpos.MakeIn(GameConst.WorldCube)
		if changed != 0 {
			//log.Printf("target %v bounce %b", o.ID, changed)
		}
	}
	lennew := a.me.PosVector.LenTo(*estpos)
	lenrate := lennew / lenori
	vt := estpos.Sub(a.me.PosVector).NormalizedTo(GameConst.MoveLimit[bulletType]).Imul(lenrate)
	return &vt
}

// --------------------- util fns ------------------

// func (a *AIAdv) calcEvasionVector(t *SPObj) *go4game.Vector3D {
// 	speed := GameConst.MoveLimit[a.me.ObjType]
// 	backvt := a.me.PosVector.Sub(t.PosVector).NormalizedTo(speed) // backward
// 	tohomevt := a.HomePos.Sub(a.me.PosVector).NormalizedTo(speed) // to home pos
// 	rtn := backvt.Add(backvt).Add(tohomevt)
// 	return &rtn
// }

func (a *AIAdv) calcAims(t *SPObj, projectilemovelimit float64) (float64, *go4game.Vector3D, float64) {
	dur := a.me.PosVector.CalcAimAheadDur(t.PosVector, t.MoveVector, projectilemovelimit)
	if math.IsInf(dur, 1) {
		return math.Inf(1), nil, 0
	}
	if t.MoveVector.Abs() < 0.1 {
		return dur, &t.PosVector, 0
	}
	estpos := t.PosVector.Add(t.MoveVector.Imul(dur))
	estangle := t.PosVector.Sub(a.me.PosVector).Angle(estpos.Sub(a.me.PosVector))
	return dur, &estpos, estangle
}
