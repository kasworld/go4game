package go4game

import (
	//"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"time"
)

type Team struct {
	ID             int
	CmdCh          chan Cmd
	Name           string
	PWorld         *World
	GameObjs       map[int]GameObject
	TeamName       string
	ClientConnInfo ConnInfo
	spp            *SpatialPartition
}

func (m Team) String() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.Name, len(m.GameObjs))
}

func NewTeam(w *World, conn net.Conn) *Team {
	t := Team{
		ID:       <-IdGenCh,
		CmdCh:    make(chan Cmd, 10),
		PWorld:   w,
		GameObjs: make(map[int]GameObject),
	}
	t.ClientConnInfo = *NewConnInfo(&t, conn)

	//log.Printf("New %v", t)
	go t.Loop()
	return &t
}

func (t *Team) Loop() {
	t.addNewGameObject()
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case cmd := <-t.CmdCh:
			switch cmd.Cmd {
			case "quit":
				t.ClientConnInfo.Conn.Close()
				break loop
			case "quitRead":
				//log.Printf("%v", cmd.Args.(error))
				t.PWorld.CmdCh <- Cmd{Cmd: "delTeam", Args: t}
				break loop
			case "quitWrite":
				//log.Printf("%v", cmd.Args.(error))
				t.PWorld.CmdCh <- Cmd{Cmd: "delTeam", Args: t}
				break loop
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		case packet := <-t.ClientConnInfo.ReadCh:
			//log.Printf("%v\n", packet)
			t.ClientConnInfo.WriteCh <- packet
			//log.Println("send/recv")
		case <-timer60Ch:
			t.spp = <-t.PWorld.SppCh
			if t.spp != nil {
				for _, v := range t.GameObjs {
					near := t.spp.GetNear2(&v.pos)
					clist := v.GetCollisionList(near)
					for _, o := range clist {
						if o.enabled {
							o.enabled = false
						}
					}
				}
				for _, v := range t.GameObjs {
					if v.enabled == false {
						t.delGameObject(&v)
					} else {
						near := t.spp.GetNear2(&v.pos)
						v.aiAction(&v, near)
						v.lastMoveTime = time.Now()
						v.curStep += 1
					}
				}
			}
			if len(t.GameObjs) < 1 {
				t.addNewGameObject()
			}
		case <-timer1secCh:
			//log.Printf("team:%v\n", t.ClientConnInfo.Stat.String())
			select {
			case t.PWorld.CmdCh <- Cmd{Cmd: "statInfo", Args: *t.ClientConnInfo.Stat}:
				t.ClientConnInfo.Stat.NewLap()
			}
		}
	}
	//log.Printf("team ending:%v\n", t.ClientConnInfo.Stat.String())
	//log.Printf("quit %v", t)
}

func (t *Team) addNewGameObject() {
	ais := []AIActionFn{autoMove1, autoMove2}
	o := *NewGameObject(t, "main", ais[rand.Intn(len(ais))])
	t.GameObjs[o.ID] = o
}

func (t *Team) delGameObject(o *GameObject) {
	delete(t.GameObjs, o.ID)
}
