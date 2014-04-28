package go4game

import (
	"fmt"
	"log"
	//"math"
	//"math/rand"
	//"net"
	//"reflect"
	"time"
)

type World struct {
	PacketStat
	ID       int
	CmdCh    chan Cmd
	PService *GameService
	MinPos   Vector3D
	MaxPos   Vector3D
	Teams    map[int]*Team
	SppCh    chan *SpatialPartition
	spp      *SpatialPartition
}

func (m World) String() string {
	return fmt.Sprintf("World:%v Team:%v", m.ID, len(m.Teams))
}

func NewWorld(g *GameService) *World {
	w := World{
		ID:         <-IdGenCh,
		PacketStat: *NewStatInfo(),
		CmdCh:      make(chan Cmd, 10),
		PService:   g,
		MinPos:     Vector3D{-500, -500, -500},
		MaxPos:     Vector3D{500, 500, 500},
		Teams:      make(map[int]*Team),
		SppCh:      make(chan *SpatialPartition),
	}
	//log.Printf("New %v", w)
	for i := 0; i < 16; i++ {
		w.addNewTeam(&AIConn{})
	}

	go w.Loop()
	return &w
}

func (w *World) SppBroadCast(CmdCh chan Cmd) {
loop:
	for {
		select {
		case <-CmdCh:
			break loop
		case w.SppCh <- w.spp:
			// broadcast
		}
	}
}

func (w *World) teamCount(ct ClientType) int {
	n := 0
	for _, t := range w.Teams {
		if t.ClientConnInfo.clientType == ct {
			n++
		}
	}
	return n
}

func (w *World) Loop() {
	SppCmdCh := make(chan Cmd, 1)
	go w.SppBroadCast(SppCmdCh)
	defer func() {
		SppCmdCh <- Cmd{Cmd: "quit"}
		for _, t := range w.Teams {
			t.CmdCh <- Cmd{Cmd: "quit"}
		}
		w.PService.CmdCh <- Cmd{Cmd: "delWorld", Args: w}
	}()

	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case cmd := <-w.CmdCh:
			switch cmd.Cmd {
			case "quit":
				break loop
			case "newTeam":
				w.addNewTeam(cmd.Args)
			case "delTeam":
				w.delTeam(cmd.Args.(*Team))
				if len(w.Teams) == w.teamCount(AIClient) {
					break loop
				}
			case "statInfo":
				s := cmd.Args.(PacketStat)
				w.PacketStat.AddLap(&s)
				//log.Println("world stat added")
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		case <-timer60Ch:
			w.spp = w.MakeSpatialPartition()
		case <-timer1secCh:
			//log.Printf("%v\n%v", w, w.PacketStat)
			select {
			case w.PService.CmdCh <- Cmd{Cmd: "statInfo", Args: w.PacketStat}:
				w.PacketStat.NewLap()
			}
		}
	}

	//log.Printf("quit %v", w)
}

func (w *World) addNewTeam(conn interface{}) {
	t := NewTeam(w, conn)
	w.Teams[t.ID] = t
}

func (w *World) delTeam(t *Team) {
	delete(w.Teams, t.ID)
}
