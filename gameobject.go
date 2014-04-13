package go4game

import (
	"fmt"
	//"math"
	//"log"
	"math/rand"
	"reflect"
	"time"
)

type GameObject struct {
	ID              int
	PTeam           *Team
	curStep         int
	enabled         bool
	objType         string
	startTime       time.Time
	endTime         time.Time
	lastMoveTime    time.Time
	MinPos          Vector3D
	MaxPos          Vector3D
	pos             Vector3D
	move            Vector3D
	accel           Vector3D
	aiAction        AIActionFn
	collisionRadius float64
	bounceDamping   float64
}

func (m GameObject) String() string {
	return fmt.Sprintf("%v ID:%v Type:%v Owner:%v",
		reflect.TypeOf(m), m.ID, m.objType, m.PTeam)
}

func NewGameObject(PTeam *Team, t string, mover AIActionFn) *GameObject {
	Min := PTeam.PWorld.MinPos
	Max := PTeam.PWorld.MaxPos
	o := GameObject{
		ID:              <-IdGenCh,
		curStep:         0,
		enabled:         true,
		objType:         t,
		startTime:       time.Now(),
		endTime:         time.Now().Add(time.Second * 60),
		lastMoveTime:    time.Now(),
		pos:             RandVector(Min, Max),
		move:            RandVector3D(-0.5, 0.5),
		accel:           RandVector3D(-0.5, 0.5),
		aiAction:        mover,
		collisionRadius: rand.Float64(),
		bounceDamping:   rand.Float64(),
		MinPos:          Min,
		MaxPos:          Max,
		PTeam:           PTeam,
	}
	//log.Printf("New %v\n", o)
	return &o
}

type GameObjectList []*GameObject

func (m *GameObject) IsCollision(target *GameObject) bool {
	return m.pos.LenTo(&target.pos) <= m.collisionRadius+target.collisionRadius
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

type AIActionFn func(m *GameObject, near GameObjectList)

func autoMove1(m *GameObject, near GameObjectList) {
	m.pos = *m.pos.Add(&m.move)
	m.move = *m.move.Add(&m.accel)
	m.move.Normalize()

	for i := range m.pos {
		if m.pos[i] > m.MaxPos[i] {
			m.pos[i] = m.MinPos[i]
		}
		if m.pos[i] < m.MinPos[i] {
			m.pos[i] = m.MaxPos[i]
		}
	}

}

func autoMove2(m *GameObject, near GameObjectList) {
	m.pos = *m.pos.Add(&m.move)
	m.move = *m.move.Add(&m.accel)

	for i := range m.pos {
		if m.pos[i] > m.MaxPos[i] {
			m.accel[i] = -m.accel[i]
			m.move[i] = -m.move[i]
			m.pos[i] = m.MaxPos[i]
		}
		if m.pos[i] < m.MinPos[i] {
			m.accel[i] = -m.accel[i]
			m.move[i] = -m.move[i]
			m.pos[i] = m.MinPos[i]
		}
	}
}
