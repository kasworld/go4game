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
	if m.spp != nil {
		return fmt.Sprintf("World%v Teams:%v spp:%v",
			m.ID, len(m.Teams), m.spp.PartCount)
	} else {
		return fmt.Sprintf("World%v Teams:%v spp:%v",
			m.ID, len(m.Teams), nil)
	}
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
	for i := 0; i < GameConst.NpcCountPerWorld; i++ {
		w.addNewTeam(&AI1{})
	}

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

func (w *World) Do1Frame(ftime time.Time) bool {
	w.updateEnv()

	for _, t := range w.Teams {
		//log.Printf("actbytime %v", t)
		t.chStep = t.doFrameWork(ftime, w.spp, w.worldSerial)
	}
	for _, t := range w.Teams {
		r, ok := <-t.chStep
		if !ok {
			t.endTeam()
			w.delTeam(t)
			if GameConst.RemoveEmptyWorld && w.teamCount(AIClient) == len(w.Teams) {
				return false
			}
		} else {
			for _, tid := range r {
				w.Teams[tid].Score += GameConst.KillScore
			}
		}
	}
	return true
}

func (w *World) Loop() {
	defer func() {
		for _, t := range w.Teams {
			t.endTeam()
			w.delTeam(t)
		}
		w.PService.CmdCh <- Cmd{Cmd: "delWorld", Args: w}
	}()

	timer60Ch := time.Tick(GameConst.FrameRate)
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
			ok := w.Do1Frame(ftime)
			if !ok {
				break loop
			}
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
