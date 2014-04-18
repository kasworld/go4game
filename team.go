package go4game

import (
	//"errors"
	"fmt"
	"log"
	//"math/rand"
	"net"
	"reflect"
	"time"
)

type Team struct {
	ID             int
	CmdCh          chan Cmd
	Name           string
	PWorld         *World
	GameObjs       map[int]*GameObject
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
		GameObjs: make(map[int]*GameObject),
	}
	t.ClientConnInfo = *NewConnInfo(&t, conn)

	//log.Printf("New %v", t)
	go t.Loop()
	return &t
}

func (t *Team) EndTeam() {
	t.ClientConnInfo.CmdCh <- Cmd{Cmd: "quit"}
	t.ClientConnInfo.Conn.Close()
	t.PWorld.CmdCh <- Cmd{Cmd: "delTeam", Args: t}
}

func (t *Team) Loop() {
	defer t.EndTeam()

	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case cmd := <-t.CmdCh:
			switch cmd.Cmd {
			case "quit", "quitRead", "quitWrite":
				break loop
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		case p := <-t.ClientConnInfo.ReadCh:
			packet := p.(GamePacket)
			switch packet.Cmd {
			case "makeTeam":
			case "action":

			}
			select {
			case t.ClientConnInfo.WriteCh <- packet:
			}

		case ftime := <-timer60Ch:
			// do automove by time
			select {
			case t.spp = <-t.PWorld.SppCh:
				if t.spp != nil {
					for _, v := range t.GameObjs {
						v.ActByTime(ftime)
						if v.enabled == false {
							t.delGameObject(v)
						}
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
	o := *NewGameObject(t, "main")
	t.GameObjs[o.ID] = &o
}

func (t *Team) delGameObject(o *GameObject) {
	delete(t.GameObjs, o.ID)
}

type AIAction struct {
}

func (t *Team) ApplyAIAction(aa *AIAction) {
	// change accel, fire bullet
}
