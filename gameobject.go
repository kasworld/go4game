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
		m.ID, m.objType, m.PTeam)
}

// func (m *GameObject) IsCollision(target *GameObject) bool {
// 	teamrule := m.PTeam != target.PTeam
// 	checklen := m.posVector.LenTo(&target.posVector) <= (m.collisionRadius + target.collisionRadius)
// 	return (teamrule) && (checklen)
// }

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
	objType   GameObjectType
	startTime time.Time
	endTime   time.Time

	MinPos             Vector3D
	MaxPos             Vector3D
	posVector          Vector3D
	moveVector         Vector3D
	moveLimit          float64
	accelVector        Vector3D
	diffToTargetVector Vector3D
	targetObjID        int
	rotateAxis         Vector3D
	rotateSpeed        float64
	bounceDamping      float64
	lastMoveTime       time.Time
	collisionRadius    float64

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
		posVector:    Vector3D{0, 0, 0},
		moveVector:   RandVector3D(-50., 50.),
		accelVector:  RandVector3D(-50., 50.),

		moveLimit:         100.0,
		bounceDamping:     1.0,
		collisionRadius:   10.,
		moveByTimeFn:      moveByTimeFn_default,
		borderActionFn:    borderActionFn_Bounce,
		collisionActionFn: collisionFn_default,
		expireActionFn:    expireFn_default,
	}
	return &o
}

func (o *GameObject) MakeMainObj() {
	o.moveLimit = 100.0
	o.collisionRadius = 10
	o.posVector = RandVector3D(-500, 500)
	o.endTime = o.startTime.Add(time.Second * 3600)
	o.objType = GameObjMain
	o.posVector[1] = 0
	o.moveVector[1] = 0
	o.accelVector[1] = 0
}
func (o *GameObject) MakeShield(mo *GameObject) {
	o.moveLimit = 200.0
	o.collisionRadius = 5
	o.endTime = o.startTime.Add(time.Second * 3600)
	o.moveByTimeFn = moveByTimeFn_shield
	o.borderActionFn = borderActionFn_None
	o.posVector = mo.posVector
	o.objType = GameObjShield
}
func (o *GameObject) MakeBullet(mo *GameObject, moveVector Vector3D) {
	o.moveLimit = 300.0
	o.collisionRadius = 5
	o.posVector = mo.posVector
	o.moveVector = moveVector
	o.borderActionFn = borderActionFn_Disable
	o.accelVector = Vector3D{0, 0, 0}
	o.objType = GameObjBullet
	o.posVector[1] = 0
	o.moveVector[1] = 0
	o.accelVector[1] = 0
}

type ActionFnEnvInfo struct {
	frameTime time.Time
}

func (o *GameObject) ActByTime(t time.Time, spp *SpatialPartition) {
	o.posVector[1] = 0
	o.moveVector[1] = 0
	o.accelVector[1] = 0

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
	if spp.IsCollision(o) {
		if o.collisionActionFn != nil {
			ok := o.collisionActionFn(o, &envInfo)
			if ok != true {
				return
			}
		}
	}
	if o.moveByTimeFn != nil {
		// change posVector by movevector
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
	if m.moveVector.Abs() > m.moveLimit {
		m.moveVector = *m.moveVector.Normalized().Imul(m.moveLimit)
	}
	dur := float64(envInfo.frameTime.Sub(m.lastMoveTime)) / float64(time.Second)
	//log.Printf("frame dur %v %v", m.lastMoveTime, dur)
	m.posVector = *m.posVector.Add(m.moveVector.Imul(dur))
	m.moveVector = *m.moveVector.Add(m.accelVector.Imul(dur))
	return true
}

func moveByTimeFn_shield(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	mo := m.PTeam.findMainObj()
	dur := float64(envInfo.frameTime.Sub(m.startTime)) / float64(time.Second)
	//axis := &Vector3D{0, math.Copysign(20, m.moveVector[0]), 0}
	axis := mo.moveVector.Normalized().Imul(20)
	//p := m.accelVector.Normalized().Imul(20)
	p := mo.moveVector.Cross(&m.moveVector).Normalized().Imul(20)
	m.posVector = *mo.posVector.Add(p.RotateAround(axis, dur))
	return true
}

func borderActionFn_Bounce(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.posVector {
		if m.posVector[i] > m.MaxPos[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.moveVector[i] = -m.moveVector[i]
			m.posVector[i] = m.MaxPos[i]
		}
		if m.posVector[i] < m.MinPos[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.moveVector[i] = -m.moveVector[i]
			m.posVector[i] = m.MinPos[i]
		}
	}
	return true
}

func borderActionFn_Wrap(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.posVector {
		if m.posVector[i] > m.MaxPos[i] {
			m.posVector[i] = m.MinPos[i]
		}
		if m.posVector[i] < m.MinPos[i] {
			m.posVector[i] = m.MaxPos[i]
		}
	}
	return true
}

func borderActionFn_Disable(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.posVector {
		if (m.posVector[i] > m.MaxPos[i]) || (m.posVector[i] < m.MinPos[i]) {
			m.enabled = false
			return false
		}
	}
	return true
}

func borderActionFn_None(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	return true
}
