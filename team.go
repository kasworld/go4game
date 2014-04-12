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
		CmdCh:          make(chan Cmd, 100),
		PWorld:         w,
		ClientConnInfo: *NewConnInfo(conn),
		GameObjs:       make(map[int]GameObject),
	}
	log.Printf("new %v\n", t.ToString())
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
		case <-timer1secCh:
			log.Printf("team:%v\n", t.ClientConnInfo.Stat.ToString())
			t.PWorld.CmdCh <- Cmd{
				Cmd:  "statInfo",
				Args: t.ClientConnInfo.Stat,
			}
			t.ClientConnInfo.Stat = *NewStatInfo()

		case <-timer60Ch:
		case packet := <-t.ClientConnInfo.ReadCh:
			t.ClientConnInfo.WriteCh <- packet
			log.Println("send/recv")
		case cmd := <-t.CmdCh:
			log.Printf("cmd %v", cmd)
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
