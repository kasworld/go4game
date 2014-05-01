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
	ID          int
	CmdCh       chan Cmd
	PService    *GameService
	MinPos      Vector3D
	MaxPos      Vector3D
	Teams       map[int]*Team
	spp         *SpatialPartition
	worldSerial *WorldSerialize
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
	}
	for i := 0; i < 1000; i++ {
		w.addNewTeam(&AIConn{})
	}

	go w.Loop()
	return &w
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

func (w *World) updateEnv() {
	chspp := make(chan bool)
	go func() {
		w.spp = w.MakeSpatialPartition()
		chspp <- true
	}()
	chwsrl := make(chan bool)
	go func() {
		w.worldSerial = NewWorldSerialize(w)
		chwsrl <- true
	}()
	<-chspp
	<-chwsrl
}

func (w *World) Loop() {
	defer func() {
		for _, t := range w.Teams {
			t.endTeam()
			w.delTeam(t)
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
			default:
				log.Printf("unknown cmd %v\n", cmd)
			}
		case ftime := <-timer60Ch:
			w.updateEnv()

			for _, t := range w.Teams {
				//log.Printf("actbytime %v", t)
				t.chStep = t.doFrameWork(ftime, w.spp, w.worldSerial)
			}
			for _, t := range w.Teams {
				//log.Printf("actbytime wait %v", t)
				r := <-t.chStep
				if !r {
					t.endTeam()
					w.delTeam(t)
				}
			}
		case <-timer1secCh:
			osum := 0
			for _, t := range w.Teams {
				osum += len(t.GameObjs)
				w.PacketStat.AddLap(t.ClientConnInfo.Stat)
				t.ClientConnInfo.Stat.NewLap()
			}
			log.Printf("%v objs:%v spp:%v ", w, osum, w.spp.PartSize)
			select {
			case w.PService.CmdCh <- Cmd{Cmd: "statInfo", Args: w.PacketStat}:
				w.PacketStat.NewLap()
			}
		}
	}
}

func (w *World) addNewTeam(conn interface{}) {
	t := NewTeam(w, conn)
	w.Teams[t.ID] = t
}

func (w *World) delTeam(t *Team) {
	delete(w.Teams, t.ID)
}
