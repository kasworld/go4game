package go4game

import (
	"math"
	//"math/rand"
	//"net"
	//"bytes"
	//"reflect"
	//"sort"
	//"text/template"
	"fmt"
	"log"
	"time"
)

type World struct {
	ID    int64
	Teams map[int64]*Team

	CmdCh           chan GoCmd
	pService        *GameService
	worldSerial     *WorldDisp
	octree          *Octree
	clientViewRange *HyperRect
}

func (w World) String() string {
	return fmt.Sprintf("World%v AIs:%v Players:%v ViewRange %v",
		w.ID, w.TeamCountByType(TeamTypeAI), w.TeamCountByType(TeamTypePlayer), w.clientViewRange)
}

func NewWorld(g *GameService) *World {
	maxclientCount := GameConst.MaxTcpClientPerWorld + GameConst.MaxWsClientPerWorld + GameConst.AICountPerWorld
	w := World{
		ID:       <-IdGenCh,
		CmdCh:    make(chan GoCmd, maxclientCount),
		pService: g,
		Teams:    make(map[int64]*Team),
	}
	return &w
}

func (w *World) addAITeamsFromString(anames []string, n int) {
	for i := 0; i < n; i++ {
		thisai := anames[i%len(anames)]
		newteam := AIstr2AIActor(thisai)
		if newteam == nil {
			continue
		}
		rsp := make(chan interface{})
		w.CmdCh <- GoCmd{Cmd: "AddTeam", Args: NewTeam(newteam, TeamTypeAI), Rsp: rsp}
		<-rsp
	}
}

func (w *World) addTeam(t *Team) {
	if w == nil {
		log.Printf("warning add team%v to nil world", t.ID)
		return
	}
	if w.Teams[t.ID] != nil {
		log.Printf("cannot add world%v team%v exist ", w.ID, t.ID)
	} else {
		//log.Printf("add team%v to world%v", t.ID, w.ID)
		w.Teams[t.ID] = t
	}
}
func (w *World) removeTeam(id int64) {
	if w == nil {
		log.Printf("warning remove team%v from nil world", id)
		return
	}
	if w.Teams[id] == nil {
		log.Printf("cannot remove world%v team%v not exist", w.ID, id)
	} else {
		//log.Printf("remove team%v from world%v", id, w.ID)
		delete(w.Teams, id)
	}
}

func (w *World) decideClientViewRange() *HyperRect {
	ocount := 0
	for _, t := range w.Teams {
		ocount += len(t.GameObjs)
	}
	n := math.Pow(float64(ocount), 1.0/3.0)
	if n < 2 {
		n = 2
	}
	hs := GameConst.WorldCube.SizeVector().Imul(1.0 / n)
	for i := 0; i < 3; i++ {
		if hs[i] < GameConst.MaxObjectRadius*3 {
			hs[i] = GameConst.MaxObjectRadius * 3
		}
	}
	hr := HyperRect{
		Min: hs.Neg(),
		Max: hs,
	}
	return &hr
}

func (w *World) updateEnv() {
	if w == nil {
		log.Printf("warning updateEnv nil world")
		return
	}

	chwsrl := make(chan *WorldDisp)
	go func() {
		chwsrl <- NewWorldDisp(w)
	}()

	chcvr := make(chan *HyperRect)
	go func() {
		chcvr <- w.decideClientViewRange()
	}()

	choctree := make(chan *Octree)
	go func() {
		choctree <- MakeOctree(w)
	}()

	w.worldSerial = <-chwsrl
	w.octree = <-choctree
	w.clientViewRange = <-chcvr
}

func (w *World) isEmpty() bool {
	return GameConst.RemoveEmptyWorld && w.TeamCountByConn(AIConn) == len(w.Teams)
}

func (w *World) Do1Frame(ftime time.Time) bool {
	w.updateEnv()
	for _, t := range w.Teams {
		t.chStep = t.Do1Frame(w, ftime)
	}
	killedTidList := make(map[int64]bool, 0)
	quitedTidList := make(map[int64]bool, 0)
	for _, t := range w.Teams {
		clist, ok := <-t.chStep
		if !ok {
			quitedTidList[t.ID] = true
		} else {
			for _, tid := range clist {
				w.Teams[tid].Score += GameConst.KillScore
			}
			if len(clist) > 0 { // main obj explode
				t.makeMainObj()
				killedTidList[t.ID] = true
			}
		}
	}
	for id, _ := range quitedTidList {
		w.Teams[id].endTeam()
		w.removeTeam(id)
		if w.isEmpty() {
			return false
		}
	}
	for id, _ := range killedTidList {
		if quitedTidList[id] { // pass quited team
			continue
		}
		nw := w.pService.nextWorld(w.ID)
		if nw == nil {
			break
		}
		//log.Printf("move team%v from world%v to world%v", id, w.ID, nw.ID)
		t := w.Teams[id]
		w.removeTeam(id)
		nw.CmdCh <- GoCmd{Cmd: "AddTeam", Args: t}
	}
	return true
}

func (w *World) Loop() {
	defer func() {
		wi := w.makeWorldInfoWeb()
		fmt.Println(wi)
		for id, t := range w.Teams {
			w.removeTeam(id)
			t.endTeam()
		}
		w.pService.CmdCh <- GoCmd{Cmd: "delWorld", Args: w}
	}()

	timer60Ch := time.Tick(time.Duration(1000/GameConst.FramePerSec) * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case cmd := <-w.CmdCh:
			//log.Printf("world%v recv cmd %v", w.ID, cmd)
			switch cmd.Cmd {
			case "quit":
				break loop
			case "AddTeam": // from world
				w.addTeam(cmd.Args.(*Team))
				if cmd.Rsp != nil {
					cmd.Rsp <- true
				}
			case "RemoveTeam": // from world
				w.removeTeam(cmd.Args.(int64))
				if cmd.Rsp != nil {
					cmd.Rsp <- true
				}
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

func (w *World) TeamCountByConn(ct ConnType) int {
	n := 0
	for _, t := range w.Teams {
		if t.ClientConnInfo != nil && t.ClientConnInfo.connType == ct {
			n++
		}
	}
	return n
}

func (w *World) TeamCountByType(tt TeamType) int {
	n := 0
	for _, t := range w.Teams {
		if t.Type == tt {
			n++
		}
	}
	return n
}
