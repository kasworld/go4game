package go4game

import (
	"fmt"
	//"math"
	//"log"
	//"math/rand"
	//"reflect"
	"time"
)

func (m GameObject) String() string {
	return fmt.Sprintf("GameObject:%v Type:%v Team%v",
		m.ID, m.ObjType, m.TeamID)
}

type GameObject struct {
	ID         int64
	TeamID     int64
	PosVector  Vector3D
	MoveVector Vector3D
	ObjType    GameObjectType

	colcount     int64
	enabled      bool
	startTime    time.Time
	endTime      time.Time
	accelVector  Vector3D
	targetObjID  int64
	targetTeamID int64
	lastMoveTime time.Time

	moveByTimeFn      GameObjectActFn
	borderActionFn    GameObjectActFn
	collisionActionFn GameObjectActFn
	expireActionFn    GameObjectActFn
}

func NewGameObject(teamID int64) *GameObject {
	o := GameObject{
		ID:                <-IdGenCh,
		TeamID:            teamID,
		enabled:           true,
		startTime:         time.Now(),
		lastMoveTime:      time.Now(),
		moveByTimeFn:      moveByTimeFn_default,
		borderActionFn:    borderActionFn_Bounce,
		collisionActionFn: collisionFn_default,
		expireActionFn:    expireFn_default,
	}
	return &o
}

func (o *GameObject) ClearY() {
	if !GameConst.ClearY {
		return
	}
	o.PosVector[1] = 0
	o.MoveVector[1] = 0
	o.accelVector[1] = 0
}

func (o *GameObject) MakeMainObj() {
	o.PosVector = GameConst.WorldCube.RandVector()
	o.MoveVector = GameConst.WorldCube.RandVector()
	o.accelVector = GameConst.WorldCube.RandVector()
	o.ObjType = GameObjMain
	o.ClearY()
}
func (o *GameObject) MakeShield(mo *GameObject) {
	o.MoveVector = GameConst.WorldCube.RandVector()
	o.accelVector = GameConst.WorldCube.RandVector()
	o.endTime = o.startTime.Add(time.Second * 60)
	o.moveByTimeFn = moveByTimeFn_shield
	o.borderActionFn = borderActionFn_None
	o.PosVector = mo.PosVector
	o.ObjType = GameObjShield
}
func (o *GameObject) MakeBullet(mo *GameObject, MoveVector Vector3D) {
	o.endTime = o.startTime.Add(time.Second * 60)
	o.PosVector = mo.PosVector
	o.MoveVector = MoveVector
	o.borderActionFn = borderActionFn_Disable
	o.accelVector = Vector3D{0, 0, 0}
	o.ObjType = GameObjBullet
	o.ClearY()
}
func (o *GameObject) MakeSuperBullet(mo *GameObject, MoveVector Vector3D) {
	o.endTime = o.startTime.Add(time.Second * 60)
	o.PosVector = mo.PosVector
	o.MoveVector = MoveVector
	o.borderActionFn = borderActionFn_Disable
	o.accelVector = Vector3D{0, 0, 0}
	o.ObjType = GameObjSuperBullet
	o.ClearY()
}
func (o *GameObject) MakeHommingBullet(mo *GameObject, targetteamid int64, targetid int64) {
	o.endTime = o.startTime.Add(time.Second * 60)
	o.PosVector = mo.PosVector
	o.borderActionFn = borderActionFn_None
	o.accelVector = Vector3D{0, 0, 0}

	o.MoveVector = Vector3D{0, 0, 0}
	o.targetObjID = targetid
	o.targetTeamID = targetteamid
	o.moveByTimeFn = moveByTimeFn_homming
	o.ObjType = GameObjHommingBullet
	o.ClearY()
}

func (o *GameObject) IsCollision(s *SPObj) bool {
	o.colcount++
	if (s.TeamID != o.TeamID) && GameConst.IsInteract[o.ObjType][s.ObjType] && (s.PosVector.Sqd(o.PosVector) <= GameConst.ObjSqd[s.ObjType][o.ObjType]) {
		return true
	}
	return false
}

type ActionFnEnvInfo struct {
	frameTime time.Time
	world     *World
}

func (o *GameObject) ActByTime(world *World, t time.Time) IDList {
	spp := world.spp

	o.ClearY()
	var clist IDList

	defer func() {
		o.lastMoveTime = t
	}()
	envInfo := ActionFnEnvInfo{
		frameTime: t,
		world:     world,
	}
	// check expire
	if !o.endTime.IsZero() && o.endTime.Before(t) {
		if o.expireActionFn != nil {
			ok := o.expireActionFn(o, &envInfo)
			if ok != true {
				return clist
			}
		}
	}
	// check if collision , disable
	// modify own status only
	var isCollsion bool
	if o.ObjType == GameObjMain {
		clist = spp.GetCollisionList(o.IsCollision, o.PosVector, GameConst.MaxObjectRadius)
		if len(clist) > 0 {
			isCollsion = true
		}
	} else {
		isCollsion = spp.IsCollision(o.IsCollision, o.PosVector, GameConst.Radius[o.ObjType]+GameConst.MaxObjectRadius)
	}
	if isCollsion {
		if o.collisionActionFn != nil {
			ok := o.collisionActionFn(o, &envInfo)
			if ok != true {
				return clist
			}
		}
	}
	if o.moveByTimeFn != nil {
		// change PosVector by movevector
		ok := o.moveByTimeFn(o, &envInfo)
		if ok != true {
			return clist
		}
	}
	if o.borderActionFn != nil {
		// check wall action ( wrap, bounce )
		ok := o.borderActionFn(o, &envInfo)
		if ok != true {
			return clist
		}
	}
	return clist
}

type GameObjectActFn func(m *GameObject, envInfo *ActionFnEnvInfo) bool

func expireFn_default(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	m.enabled = false
	return false
}

func collisionFn_default(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	m.enabled = false
	return false
}

func moveByTimeFn_default(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	dur := float64(envInfo.frameTime.Sub(m.lastMoveTime)) / float64(time.Second)
	//log.Printf("frame dur %v %v", m.lastMoveTime, dur)
	m.MoveVector = m.MoveVector.Add(m.accelVector.Imul(dur))
	if m.MoveVector.Abs() > GameConst.MoveLimit[m.ObjType] {
		m.MoveVector = m.MoveVector.NormalizedTo(GameConst.MoveLimit[m.ObjType])
	}
	m.PosVector = m.PosVector.Add(m.MoveVector.Imul(dur))
	return true
}

func moveByTimeFn_shield(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	mo := envInfo.world.Teams[m.TeamID].findMainObj()
	if mo == nil {
		return false
	}
	dur := float64(envInfo.frameTime.Sub(m.startTime)) / float64(time.Second)
	//axis := &Vector3D{0, math.Copysign(20, m.MoveVector[0]), 0}
	axis := mo.MoveVector.NormalizedTo(20)
	//p := m.accelVector.NormalizedTo(20)
	p := mo.MoveVector.Cross(m.MoveVector).NormalizedTo(20)
	m.PosVector = mo.PosVector.Add(p.RotateAround(axis, dur+m.accelVector.Abs()))
	return true
}

func moveByTimeFn_homming(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	// how to other team obj pos? without panic
	targetTeam := envInfo.world.Teams[m.targetTeamID]
	if targetTeam == nil {
		m.enabled = false
		return false
	}
	targetobj := targetTeam.GameObjs[m.targetObjID]
	if targetobj == nil {
		m.enabled = false
		return false
	}
	m.accelVector = targetobj.PosVector.Sub(m.PosVector).NormalizedTo(GameConst.MoveLimit[m.ObjType])
	return moveByTimeFn_default(m, envInfo)
}

func borderActionFn_Bounce(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > GameConst.WorldCube.Max[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.MoveVector[i] = -m.MoveVector[i]
			m.PosVector[i] = GameConst.WorldCube.Max[i]
		}
		if m.PosVector[i] < GameConst.WorldCube.Min[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.MoveVector[i] = -m.MoveVector[i]
			m.PosVector[i] = GameConst.WorldCube.Min[i]
		}
	}
	return true
}

func borderActionFn_Wrap(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > GameConst.WorldCube.Max[i] {
			m.PosVector[i] = GameConst.WorldCube.Min[i]
		}
		if m.PosVector[i] < GameConst.WorldCube.Min[i] {
			m.PosVector[i] = GameConst.WorldCube.Max[i]
		}
	}
	return true
}

func borderActionFn_Disable(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if (m.PosVector[i] > GameConst.WorldCube.Max[i]) || (m.PosVector[i] < GameConst.WorldCube.Min[i]) {
			m.enabled = false
			return false
		}
	}
	return true
}

func borderActionFn_None(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	return true
}
