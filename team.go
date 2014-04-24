package go4game

import (
	//"errors"
	"fmt"
	"log"
	//"math/rand"
	"net"
	//"reflect"
	"github.com/gorilla/websocket"
	"time"
)

/*
  non client team : == serverside ai
  tcp client team
  web client team : websocket
*/

type Team struct {
	ID             int
	CmdCh          chan Cmd
	PWorld         *World
	GameObjs       map[int]*GameObject
	ClientConnInfo ConnInfo
	spp            *SpatialPartition
}

func (m Team) String() string {
	return fmt.Sprintf("Team:%v Obj:%v", m.ID, len(m.GameObjs))
}

func NewTeam(w *World, conn interface{}) *Team {
	t := Team{
		ID:       <-IdGenCh,
		CmdCh:    make(chan Cmd, 10),
		PWorld:   w,
		GameObjs: make(map[int]*GameObject),
	}
	switch conn.(type) {
	case net.Conn:
		t.ClientConnInfo = *NewConnInfo(&t, conn.(net.Conn))
	case *websocket.Conn:
		t.ClientConnInfo = *NewWsConnInfo(&t, conn.(*websocket.Conn))
	default:
		log.Printf("unknown type %#v", conn)
	}
	go t.Loop()
	return &t
}

func (t *Team) Loop() {
	defer func() {
		close(t.ClientConnInfo.WriteCh) // stop writeloop
		if t.ClientConnInfo.Conn != nil {
			t.ClientConnInfo.Conn.Close() // stop read loop
		}
		if t.ClientConnInfo.WsConn != nil {
			t.ClientConnInfo.WsConn.Close() // stop read loop
		}
		t.PWorld.CmdCh <- Cmd{Cmd: "delTeam", Args: t}
	}()

	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case cmd := <-t.CmdCh:
			switch cmd.Cmd {
			case "quit":
				break loop
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		case p, ok := <-t.ClientConnInfo.ReadCh:
			if !ok { // read closed
				break loop
			}
			switch p.Cmd {
			case ReqMakeTeam:
				rp := GamePacket{
					Cmd: RspMakeTeam,
				}
				t.ClientConnInfo.WriteCh <- &rp
			case ReqWorldInfo:
				rp := GamePacket{
					Cmd:       RspWorldInfo,
					WorldInfo: NewWorldSerialize(t.PWorld),
				}
				t.ClientConnInfo.WriteCh <- &rp
			case ReqAIAct:
				rp := GamePacket{
					Cmd: RspAIAct,
				}
				t.ClientConnInfo.WriteCh <- &rp
			default:
				log.Printf("unknown packet %#v", p)
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
	//log.Printf("quit %v\n", t)
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
