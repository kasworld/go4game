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
	ID              int
	CmdCh           chan Cmd
	PService        *GameService
	MinPos          Vector3D
	MaxPos          Vector3D
	Teams           map[int]*Team
	spp             *SpatialPartition
	worldSerial     *WorldSerialize
	MaxObjectRadius float64
}

func (m World) String() string {
	return fmt.Sprintf("World%v Teams:%v",
		m.ID, len(m.Teams))
}

func NewWorld(g *GameService) *World {
	w := World{
		ID:              <-IdGenCh,
		CmdCh:           make(chan Cmd, 10),
		PService:        g,
		MinPos:          GameConst.WorldMin,
		MaxPos:          GameConst.WorldMax,
		Teams:           make(map[int]*Team),
		MaxObjectRadius: GameConst.MaxObjectRadius,
	}
	for i := 0; i < w.PService.config.NpcCountPerWorld; i++ {
		w.addNewTeam(&AIConn{})
	}

	//go w.Loop()
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

func (w *World) Do1Frame(ftime time.Time) {
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
	// if w.teamCount(AIClient) == len(w.Teams) {
	// 	break loop
	// }
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
			w.Do1Frame(ftime)
		case <-timer1secCh:
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
