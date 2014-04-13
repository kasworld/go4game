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
	SppCh    chan *SpatialPartition
	spp      *SpatialPartition
}

func (m World) String() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.Name, len(m.Teams))
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
		SppCh:    make(chan *SpatialPartition),
	}
	log.Printf("New %v", w)
	go w.Loop()
	return &w
}

func (w *World) SppBroadCast() {
	for {
		w.SppCh <- w.spp
	}
}

func (w *World) Loop() {
	go w.SppBroadCast()
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case <-timer1secCh:
			log.Printf("world:%v, team:%v",
				w.StatInfo.String(),
				len(w.Teams),
			)
		case <-timer60Ch:
			w.spp = w.MakeSpatialPartition()
			// for _, t := range w.Teams {
			// 	t.CmdCh <- Cmd{
			// 		Cmd:  "envInfo",
			// 		Args: spp,
			// 	}
			// }
		case w.PService.CmdCh <- Cmd{Cmd: "statInfo", Args: &w.StatInfo}:
			w.StatInfo.NewLap()
		case cmd := <-w.CmdCh:
			switch cmd.Cmd {
			case "quit":
				for _, t := range w.Teams {
					t.CmdCh <- Cmd{Cmd: "quit"}
				}
				break loop
			case "newTeam":
				w.addNewTeam(cmd.Args.(net.Conn))
			case "delTeam":
				w.delTeam(cmd.Args.(*Team))
			case "statInfo":
				w.StatInfo.AddLap(cmd.Args.(*StatInfo))
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

func (w *World) delTeam(t *Team) {
	delete(w.Teams, t.ID)
}
