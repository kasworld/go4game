package go4game

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"time"
)

type Team struct {
	StatInfo
	ID             int
	CmdCh          chan Cmd
	Name           string
	PWorld         *World
	GameObjs       map[int]GameObject
	TeamName       string
	ClientConnInfo ConnInfo
}

func (m *Team) ToString() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.Name, len(m.GameObjs))
}

func NewTeam(w *World, conn net.Conn) *Team {
	t := Team{
		ID:             <-IdGenCh,
		StatInfo:       *NewStatInfo(),
		CmdCh:          make(chan Cmd),
		PWorld:         w,
		ClientConnInfo: *NewConnInfo(conn),
		GameObjs:       make(map[int]GameObject),
	}
	t.addNewGameObject()
	log.Printf("new %v", t.ToString())
	go t.Loop()
	return &t
}

func (t *Team) Loop() {
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case <-timer1secCh:
			t.PWorld.CmdCh <- Cmd{
				Cmd:  "info",
				Args: t.StatInfo,
			}
			fmt.Printf("world:%v\n", t.StatInfo.ToString())
			t.StatInfo = *NewStatInfo()

		case <-timer60Ch:
		case packet := <-t.ClientConnInfo.ReadCh:
			t.ClientConnInfo.WriteCh <- packet
		case cmd := <-t.CmdCh:
			switch cmd.Cmd {
			case "quit":
				for _, v := range t.GameObjs {
					v.CmdCh <- Cmd{Cmd: "quit"}
				}
				t.ClientConnInfo.Conn.Close()
				break loop
			case "envInfo":
				for _, v := range t.GameObjs {
					v.CmdCh <- cmd
				}
			case "attacked":
				var attacked *GameObject = cmd.Args.(*GameObject)
				t.delGameObject(attacked)
				if len(t.GameObjs) < 1 {
					t.addNewGameObject()
				}
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		}
	}
}

func (t *Team) addNewGameObject() {
	ais := []AIActionFn{autoMove1, autoMove2}
	o := *NewGameObject(t, "main", ais[rand.Intn(len(ais))])
	t.GameObjs[o.ID] = o
}

func (t *Team) delGameObject(o *GameObject) {
	delete(t.GameObjs, o.ID)
}
