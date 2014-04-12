package go4game

import (
	"fmt"
	//"math"
	"log"
	"math/rand"
	"reflect"
	"time"
)

type GameObject struct {
	ID              int
	CmdCh           chan Cmd
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

func NewGameObject(PTeam *Team, t string, mover AIActionFn) *GameObject {
	Min := PTeam.PWorld.MinPos
	Max := PTeam.PWorld.MaxPos
	o := GameObject{
		ID:              <-IdGenCh,
		CmdCh:           make(chan Cmd, 100),
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
	}
	log.Printf("new %v", o.ToString())
	go o.Loop()
	return &o
}

func (m *GameObject) Loop() {
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case <-timer1secCh:
		case <-timer60Ch:
		case cmd := <-m.CmdCh:
			switch cmd.Cmd {
			case "quit":
				break loop
			case "envInfo":
				var spp *SpatialPartition = cmd.Args.(*SpatialPartition)
				near := spp.GetNear2(&m.pos)
				clist := m.GetCollisionList(near)
				for _, o := range clist {
					o.CmdCh <- Cmd{
						Cmd:  "attackedBy",
						Args: m,
					}
				}
				m.aiAction(m, near)
				m.lastMoveTime = time.Now()
				m.curStep += 1

			case "attackedBy":
				//var attacker *GameObject = cmd.Args.(*GameObject)
				m.enabled = false
				m.PTeam.CmdCh <- Cmd{
					Cmd:  "attacked",
					Args: m,
				}
				break loop
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}

		}
	}
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

func (m *GameObject) ToString() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.pos, m.curStep)
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
