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

const (
	_ = iota
	GameObjMain
	GameObjShield
	GameObjBullet
)

type GameObjectType int

type GameObject struct {
	ID        int
	PTeam     *Team
	enabled   bool
	ObjType   GameObjectType
	startTime time.Time
	endTime   time.Time

	MinPos             Vector3D
	MaxPos             Vector3D
	PosVector          Vector3D
	MoveVector         Vector3D
	moveLimit          float64
	accelVector        Vector3D
	diffToTargetVector Vector3D
	targetObjID        int
	rotateAxis         Vector3D
	rotateSpeed        float64
	bounceDamping      float64
	lastMoveTime       time.Time
	CollisionRadius    float64

	moveByTimeFn      GameObjectActFn
	borderActionFn    GameObjectActFn
	collisionActionFn GameObjectActFn
	expireActionFn    GameObjectActFn
}

func NewGameObject(PTeam *Team) *GameObject {
	Min := PTeam.PWorld.MinPos
	Max := PTeam.PWorld.MaxPos
	o := GameObject{
		ID:        <-IdGenCh,
		PTeam:     PTeam,
		enabled:   true,
		startTime: time.Now(),
		endTime:   time.Now().Add(time.Second * 60),

		lastMoveTime: time.Now(),
		MinPos:       Min,
		MaxPos:       Max,
		PosVector:    Vector3D{0, 0, 0},
		MoveVector:   *RandVector3D(-500., 500.),
		accelVector:  *RandVector3D(-500., 500.),

		moveLimit:         100.0,
		bounceDamping:     1.0,
		CollisionRadius:   10.,
		moveByTimeFn:      moveByTimeFn_default,
		borderActionFn:    borderActionFn_Bounce,
		collisionActionFn: collisionFn_default,
		expireActionFn:    expireFn_default,
	}
	return &o
}

func (o *GameObject) ClearY() {
	//return
	o.PosVector[1] = 0
	o.MoveVector[1] = 0
	o.accelVector[1] = 0
}

func (o *GameObject) MakeMainObj() {
	o.moveLimit = 100.0
	o.CollisionRadius = 10
	o.PosVector = *RandVector3D(-500, 500)
	o.endTime = o.startTime.Add(time.Second * 3600)
	o.ObjType = GameObjMain

	o.ClearY()
}
func (o *GameObject) MakeShield(mo *GameObject) {
	o.moveLimit = 200.0
	o.CollisionRadius = 5
	o.endTime = o.startTime.Add(time.Second * 3600)
	o.moveByTimeFn = moveByTimeFn_shield
	o.borderActionFn = borderActionFn_None
	o.PosVector = mo.PosVector
	o.ObjType = GameObjShield
}
func (o *GameObject) MakeBullet(mo *GameObject, MoveVector *Vector3D) {
	o.moveLimit = 300.0
	o.CollisionRadius = 5
	o.PosVector = mo.PosVector
	o.MoveVector = *MoveVector
	o.borderActionFn = borderActionFn_Disable
	o.accelVector = Vector3D{0, 0, 0}
	o.ObjType = GameObjBullet

	o.ClearY()
}

type ActionFnEnvInfo struct {
	frameTime time.Time
}

func (o *GameObject) IsCollision(sl SPObjList) bool {
	for _, s := range sl {
		teamrule := s.TeamID != o.PTeam.ID
		checklen := s.PosVector.LenTo(&o.PosVector) <= (s.CollisionRadius + o.CollisionRadius)
		if (teamrule) && (checklen) {
			return true
		}
	}
	return false
}

func (o *GameObject) ActByTime(t time.Time, spp *SpatialPartition) {
	o.ClearY()

	defer func() {
		o.lastMoveTime = t
	}()
	envInfo := ActionFnEnvInfo{
		frameTime: t,
	}
	// check expire
	if o.endTime.Before(t) {
		if o.expireActionFn != nil {
			ok := o.expireActionFn(o, &envInfo)
			if ok != true {
				return
			}
		}
	}
	// check if collision , disable
	// modify own status only
	//if spp.ApplyCollisionAction4(IsCollision, o) {
	if spp.ApplyPartsFn(o.IsCollision, o.PosVector, spp.MaxObjectRadius) {
		if o.collisionActionFn != nil {
			ok := o.collisionActionFn(o, &envInfo)
			if ok != true {
				return
			}
		}
	}
	if o.moveByTimeFn != nil {
		// change PosVector by movevector
		ok := o.moveByTimeFn(o, &envInfo)
		if ok != true {
			return
		}

	}
	if o.borderActionFn != nil {
		// check wall action ( wrap, bounce )
		ok := o.borderActionFn(o, &envInfo)
		if ok != true {
			return
		}
	}
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
	if m.MoveVector.Abs() > m.moveLimit {
		m.MoveVector = *m.MoveVector.Normalized().Imul(m.moveLimit)
	}
	dur := float64(envInfo.frameTime.Sub(m.lastMoveTime)) / float64(time.Second)
	//log.Printf("frame dur %v %v", m.lastMoveTime, dur)
	m.PosVector = *m.PosVector.Add(m.MoveVector.Imul(dur))
	m.MoveVector = *m.MoveVector.Add(m.accelVector.Imul(dur))
	return true
}

func moveByTimeFn_shield(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	mo := m.PTeam.findMainObj()
	dur := float64(envInfo.frameTime.Sub(m.startTime)) / float64(time.Second)
	//axis := &Vector3D{0, math.Copysign(20, m.MoveVector[0]), 0}
	axis := mo.MoveVector.Normalized().Imul(20)
	//p := m.accelVector.Normalized().Imul(20)
	p := mo.MoveVector.Cross(&m.MoveVector).Normalized().Imul(20)
	m.PosVector = *mo.PosVector.Add(p.RotateAround(axis, dur))
	return true
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
