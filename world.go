package go4game

import (
	"fmt"
	"log"
	//"math"
	//"math/rand"
	//"net"
	//"reflect"
	//"html/template"
	"sort"
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

type WorldInfo struct {
	Disp  string
	Teams []TeamInfo
}

func (m *World) makeWorldInfo() *WorldInfo {
	rtn := &WorldInfo{
		Disp:  m.String(),
		Teams: make([]TeamInfo, 0, len(m.Teams)),
	}
	for _, t := range m.Teams {
		rtn.Teams = append(rtn.Teams, *t.NewTeamInfo())
	}
	sort.Sort(ByScore(rtn.Teams))
	return rtn
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
	for i := 0; i < GameConst.NpcCountPerWorld/4; i++ {
		// w.addNewTeam(&AINothing{})
		w.addNewTeam(&AICloud{})
		w.addNewTeam(&AIRandom{})
		w.addNewTeam(&AI2{})
		w.addNewTeam(&AI3{})
	}

	return &w
}
func (w *World) addNewTeam(conn interface{}) {
	t := NewTeam(w, conn)
	w.Teams[t.ID] = t
}

func (w *World) updateEnv() {
	chspp := make(chan *SpatialPartition)
	go func() {
		chspp <- w.MakeSpatialPartition()
	}()
	chwsrl := make(chan *WorldSerialize)
	go func() {
		chwsrl <- NewWorldSerialize(w)
	}()
	w.spp = <-chspp
	w.worldSerial = <-chwsrl
}

func (w *World) Do1Frame(ftime time.Time) bool {
	w.updateEnv()

	for _, t := range w.Teams {
		t.chStep = t.doFrameWork(ftime, w.spp, w.worldSerial)
	}
	for id, t := range w.Teams {
		r, ok := <-t.chStep
		if !ok {
			t.endTeam()
			delete(w.Teams, id)
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
		for id, t := range w.Teams {
			t.endTeam()
			delete(w.Teams, id)
		}
		w.PService.CmdCh <- Cmd{Cmd: "delWorld", Args: w}
	}()

	timer60Ch := time.Tick(time.Duration(1000/GameConst.FramePerSec) * time.Millisecond)
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
			//log.Printf("in frame %v %v", ftime, w)
			ok := w.Do1Frame(ftime)
			if !ok {
				break loop
			}
			//log.Printf("out frame %v %v", ftime, w)
		case <-timer1secCh:
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
