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
	return fmt.Sprintf("GameObject:%v Type:%v Owner:%v",
		m.ID, m.ObjType, m.PTeam)
}

type GameObject struct {
	ID        int64
	PTeam     *Team
	enabled   bool
	ObjType   GameObjectType
	startTime time.Time
	endTime   time.Time

	MinPos          Vector3D
	MaxPos          Vector3D
	PosVector       Vector3D
	MoveVector      Vector3D
	moveLimit       float64
	accelVector     Vector3D
	targetObjID     int64
	targetTeamID    int64
	rotateAxis      Vector3D
	rotateSpeed     float64
	bounceDamping   float64
	lastMoveTime    time.Time
	CollisionRadius float64

	moveByTimeFn      GameObjectActFn
	borderActionFn    GameObjectActFn
	collisionActionFn GameObjectActFn
	expireActionFn    GameObjectActFn
}

func NewGameObject(PTeam *Team) *GameObject {
	Min := PTeam.PWorld.MinPos
	Max := PTeam.PWorld.MaxPos
	o := GameObject{
		ID:                <-IdGenCh,
		PTeam:             PTeam,
		enabled:           true,
		startTime:         time.Now(),
		lastMoveTime:      time.Now(),
		MinPos:            Min,
		MaxPos:            Max,
		bounceDamping:     1.0,
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
	o.PosVector = RandVector(o.MinPos, o.MaxPos)
	o.MoveVector = RandVector(o.MinPos, o.MaxPos)
	o.accelVector = RandVector(o.MinPos, o.MaxPos)

	o.ObjType = GameObjMain
	o.moveLimit = ObjDefault.MoveLimit[o.ObjType]
	o.CollisionRadius = ObjDefault.Radius[o.ObjType]

	o.ClearY()
}
func (o *GameObject) MakeShield(mo *GameObject) {
	o.MoveVector = RandVector(o.MinPos, o.MaxPos)
	o.accelVector = RandVector(o.MinPos, o.MaxPos)
	o.endTime = o.startTime.Add(time.Second * 60)
	o.moveByTimeFn = moveByTimeFn_shield
	o.borderActionFn = borderActionFn_None
	o.PosVector = mo.PosVector

	o.ObjType = GameObjShield
	o.moveLimit = ObjDefault.MoveLimit[o.ObjType]
	o.CollisionRadius = ObjDefault.Radius[o.ObjType]
}
func (o *GameObject) MakeBullet(mo *GameObject, MoveVector Vector3D) {
	o.endTime = o.startTime.Add(time.Second * 60)
	o.PosVector = mo.PosVector
	o.MoveVector = MoveVector
	o.borderActionFn = borderActionFn_Disable
	o.accelVector = Vector3D{0, 0, 0}

	o.ObjType = GameObjBullet
	o.moveLimit = ObjDefault.MoveLimit[o.ObjType]
	o.CollisionRadius = ObjDefault.Radius[o.ObjType]

	o.ClearY()
}
func (o *GameObject) MakeSuperBullet(mo *GameObject, MoveVector Vector3D) {
	o.endTime = o.startTime.Add(time.Second * 60)
	o.PosVector = mo.PosVector
	o.MoveVector = MoveVector
	o.borderActionFn = borderActionFn_Disable
	o.accelVector = Vector3D{0, 0, 0}

	o.ObjType = GameObjSuperBullet
	o.moveLimit = ObjDefault.MoveLimit[o.ObjType]
	o.CollisionRadius = ObjDefault.Radius[o.ObjType]

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
	o.moveLimit = ObjDefault.MoveLimit[o.ObjType]
	o.CollisionRadius = ObjDefault.Radius[o.ObjType]

	o.ClearY()
}

type ActionFnEnvInfo struct {
	frameTime time.Time
}

func (o *GameObject) IsCollision(s *SPObj) bool {
	o.PTeam.CollisionStat.Inc()
	if (s.TeamID != o.PTeam.ID) && InteractionMap[o.ObjType][s.ObjType] && (s.PosVector.Sqd(o.PosVector) <= ObjSqd[s.ObjType][o.ObjType]) {
		return true
	}
	return false
}

func (o *GameObject) ActByTime(t time.Time, spp *SpatialPartition) IDList {
	o.ClearY()
	var clist IDList

	defer func() {
		o.lastMoveTime = t
	}()
	envInfo := ActionFnEnvInfo{
		frameTime: t,
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
		clist = spp.GetCollisionList(o.IsCollision, o.PosVector, spp.MaxObjectRadius)
		if len(clist) > 0 {
			isCollsion = true
		}
	} else {
		isCollsion = spp.IsCollision(o.IsCollision, o.PosVector, o.CollisionRadius+spp.MaxObjectRadius)
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
	if m.MoveVector.Abs() > m.moveLimit {
		m.MoveVector = m.MoveVector.Normalized().Imul(m.moveLimit)
	}
	m.PosVector = m.PosVector.Add(m.MoveVector.Imul(dur))
	return true
}

func moveByTimeFn_shield(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	mo := m.PTeam.findMainObj()
	if mo == nil {
		return false
	}
	dur := float64(envInfo.frameTime.Sub(m.startTime)) / float64(time.Second)
	//axis := &Vector3D{0, math.Copysign(20, m.MoveVector[0]), 0}
	axis := mo.MoveVector.Normalized().Imul(20)
	//p := m.accelVector.Normalized().Imul(20)
	p := mo.MoveVector.Cross(m.MoveVector).Normalized().Imul(20)
	m.PosVector = mo.PosVector.Add(p.RotateAround(axis, dur+m.accelVector.Abs()))
	return true
}

func moveByTimeFn_homming(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	targetobj := m.PTeam.PWorld.Teams[m.targetTeamID].GameObjs[m.targetObjID]
	if targetobj == nil {
		m.enabled = false
		return false
	}
	m.accelVector = targetobj.PosVector.Sub(m.PosVector).NormalizedTo(m.moveLimit)
	return moveByTimeFn_default(m, envInfo)
}

func borderActionFn_Bounce(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > m.MaxPos[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.MoveVector[i] = -m.MoveVector[i]
			m.PosVector[i] = m.MaxPos[i]
		}
		if m.PosVector[i] < m.MinPos[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.MoveVector[i] = -m.MoveVector[i]
			m.PosVector[i] = m.MinPos[i]
		}
	}
	return true
}

func borderActionFn_Wrap(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > m.MaxPos[i] {
			m.PosVector[i] = m.MinPos[i]
		}
		if m.PosVector[i] < m.MinPos[i] {
			m.PosVector[i] = m.MaxPos[i]
		}
	}
	return true
}

func borderActionFn_Disable(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if (m.PosVector[i] > m.MaxPos[i]) || (m.PosVector[i] < m.MinPos[i]) {
			m.enabled = false
			return false
		}
	}
	return true
}

func borderActionFn_None(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	return true
}
