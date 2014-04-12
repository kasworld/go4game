package go4game

import (
	"fmt"
	"log"
	//"math"
	//"math/rand"
	"net"
	"reflect"
	"time"
)

type World struct {
	StatInfo
	ID       int
	CmdCh    chan Cmd
	PService *GameService
	Name     string
	MinPos   Vector3D
	MaxPos   Vector3D
	Teams    map[int]Team
}

func NewWorld(g *GameService) *World {
	w := World{
		ID:       <-IdGenCh,
		StatInfo: *NewStatInfo(),
		CmdCh:    make(chan Cmd),
		PService: g,
		MinPos:   Vector3D{0, 0, 0},
		MaxPos:   Vector3D{1000, 1000, 1000},
		Teams:    make(map[int]Team),
	}
	log.Printf("new %v", w.ToString())
	go w.Loop()
	return &w
}

func (w *World) Loop() {
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case <-timer1secCh:
			log.Printf("world:%v, team:%v",
				w.StatInfo.ToString(),
				len(w.Teams),
			)
		case w.PService.CmdCh <- Cmd{Cmd: "statInfo", Args: &w.StatInfo}:
			w.StatInfo.Reset()

		case <-timer60Ch:
			spp := w.MakeSpatialPartition()
			for _, t := range w.Teams {
				t.CmdCh <- Cmd{
					Cmd:  "envInfo",
					Args: spp,
				}
			}
		case cmd := <-w.CmdCh:
			switch cmd.Cmd {
			case "quit":
				for _, v := range w.Teams {
					v.CmdCh <- Cmd{Cmd: "quit"}
				}
				break loop
			case "newTeam":
				w.addNewTeam(cmd.Args.(net.Conn))
			case "statInfo":
				w.StatInfo.Add(cmd.Args.(*StatInfo))
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		}
	}
}

func (w *World) addNewTeam(conn net.Conn) {
	t := *NewTeam(w, conn)
	w.Teams[t.ID] = t
}

func (w *World) delWorld(t *Team) {
	delete(w.Teams, t.ID)
}

func (m *World) ToString() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.Name, len(m.Teams))
}
