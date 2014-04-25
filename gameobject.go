package go4game

import (
	"fmt"
	//"math"
	"log"
	//"math/rand"
	//"reflect"
	"time"
)

/*
order of process
AI make ai action
aiaction : change accelvector , fire(==req to make new object) ,
movefn change movevector ( by type)
automovebytime change pos (by time.duration)
checkwall action change pos,move,accel by Min/Max pos
check collision and disable object (in world action)
remove object ( in world )

object event
automovebytime
wallcollisionaction
collisionAction
exfireFn

*/

func (m GameObject) String() string {
	return fmt.Sprintf("GameObject:%v Type:%v Owner:%v",
		m.ID, m.objType, m.PTeam)
}

func (m *GameObject) IsCollision(target *GameObject) bool {
	return m.posVector.LenTo(&target.posVector) <= m.collisionRadius+target.collisionRadius
}

func (m *GameObject) GetCollisionList(near GameObjectList) GameObjectList {
	rtn := GameObjectList{}
	for _, o := range near {
		if m != o && m.IsCollision(o) {
			rtn = append(rtn, o)
		}
	}
	return rtn
}

type GameObject struct {
	ID        int
	PTeam     *Team
	enabled   bool
	objType   string
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

func NewGameObject(PTeam *Team, t string) *GameObject {
	Min := PTeam.PWorld.MinPos
	Max := PTeam.PWorld.MaxPos
	o := GameObject{
		ID:        <-IdGenCh,
		PTeam:     PTeam,
		enabled:   true,
		objType:   t,
		startTime: time.Now(),
		endTime:   time.Now().Add(time.Second * 60),

		lastMoveTime: time.Now(),
		MinPos:       Min,
		MaxPos:       Max,
		posVector:    RandVector(Min, Max),
		moveVector:   RandVector3D(-0.5, 0.5),
		accelVector:  RandVector3D(-0.5, 0.5),

		moveLimit:         1.0,
		bounceDamping:     1.0,
		collisionRadius:   0.1,
		moveByTimeFn:      moveByTimeFn_default,
		borderActionFn:    borderActionFn_Bounce,
		collisionActionFn: collisionFn_default,
		expireActionFn:    expireFn_default,
	}
	//log.Printf("New %v\n", o)
	return &o
}

func NewGameObject_main(PTeam *Team) *GameObject {
	Min := PTeam.PWorld.MinPos
	Max := PTeam.PWorld.MaxPos
	o := GameObject{
		ID:        <-IdGenCh,
		PTeam:     PTeam,
		enabled:   true,
		objType:   "main",
		startTime: time.Now(),
		endTime:   time.Now().Add(time.Second * 60),

		lastMoveTime: time.Now(),
		MinPos:       Min,
		MaxPos:       Max,
		posVector:    RandVector(Min, Max),
		moveVector:   RandVector3D(-0.5, 0.5),
		accelVector:  RandVector3D(-0.5, 0.5),

		moveLimit:         1.0,
		bounceDamping:     1.0,
		collisionRadius:   0.1,
		moveByTimeFn:      moveByTimeFn_default,
		borderActionFn:    borderActionFn_Bounce,
		collisionActionFn: collisionFn_default,
		expireActionFn:    expireFn_default,
	}
	log.Printf("New %v\n", o)
	return &o
}

type ActionFnEnvInfo struct {
	spp       *SpatialPartition
	near      GameObjectList
	clist     GameObjectList
	frameTime time.Time
}

func (o *GameObject) ActByTime(t time.Time) {
	defer func(o *GameObject, t time.Time) {
		o.lastMoveTime = t
	}(o, t)
	envInfo := ActionFnEnvInfo{
		spp:       o.PTeam.spp,
		near:      o.PTeam.spp.GetNear2(&o.posVector),
		frameTime: t,
	}
	envInfo.clist = o.GetCollisionList(envInfo.near)
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
	if len(envInfo.clist) > 0 {
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

type GameObjectList []*GameObject

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

func borderActionFn_None(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.posVector {
		if (m.posVector[i] > m.MaxPos[i]) || (m.posVector[i] < m.MinPos[i]) {
			m.enabled = false
			return false
		}
	}
	return true
}
