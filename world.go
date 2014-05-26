package go4game

import (
	"fmt"
	"log"
	//"math"
	//"math/rand"
	//"net"
	"bytes"
	//"reflect"
	"sort"
	"text/template"
	"time"
)

type World struct {
	ID          int64
	CmdCh       chan Cmd
	PService    *GameService
	Teams       map[int64]*Team
	spp         *SpatialPartition
	worldSerial *WorldDisp
}

func (m World) String() string {
	if m.spp != nil {
		return fmt.Sprintf("World%v Teams:%v spp:%v objs:%v",
			m.ID, len(m.Teams), m.spp.PartCount, m.spp.ObjectCount)
	} else {
		return fmt.Sprintf("World%v Teams:%v spp:%v",
			m.ID, len(m.Teams), nil)
	}
}

var WorldTextTemplate *template.Template

func init() {
	const tworld = `
{{.Disp}}
TeamColor TeamID ClientInfo ObjCount Score ActionPoint PacketStat CollStat {{range .Teams}}
{{.FontColor | printf "%x"}} {{.ID}} {{.ClientInfo}} {{.Objs}} {{.Score}} {{.AP}} {{.PacketStat}} {{.CollStat}} {{end}}
`
	WorldTextTemplate = template.Must(template.New("indexpage").Parse(tworld))
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

func (wi WorldInfo) String() string {
	var w bytes.Buffer
	WorldTextTemplate.Execute(&w, wi)
	return w.String()
}

func NewWorld(g *GameService) *World {
	w := World{
		ID:       <-IdGenCh,
		CmdCh:    make(chan Cmd, 1),
		PService: g,
		Teams:    make(map[int64]*Team),
	}
	return &w
}

func (w *World) addAITeams(anames []string, n int) {
	for i := 0; i < n; i++ {
		thisai := anames[i%len(anames)]
		switch thisai {
		default:
			log.Printf("unknown AI %v", thisai)
		case "AINothing":
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(&AINothing{})}
		case "AICloud":
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(&AICloud{})}
		case "AIRandom":
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(&AIRandom{})}
		case "AI2":
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(&AI2{})}
		case "AI3":
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(&AI3{})}
		}
	}
}

func (w *World) addTeam(t *Team) {
	w.Teams[t.ID] = t
}
func (w *World) removeTeam(id int64) {
	delete(w.Teams, id)
}

func (w *World) updateEnv() {
	chspp := make(chan *SpatialPartition)
	go func() {
		chspp <- w.MakeSpatialPartition()
	}()
	chwsrl := make(chan *WorldDisp)
	go func() {
		chwsrl <- NewWorldDisp(w)
	}()
	w.spp = <-chspp
	w.worldSerial = <-chwsrl
}

func (w *World) Do1Frame(ftime time.Time) bool {
	w.updateEnv()

	for _, t := range w.Teams {
		t.chStep = t.doFrameWork(w, ftime)
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
		wi := w.makeWorldInfo()
		fmt.Println(wi)
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
			case "AddTeam": // from world
				w.addTeam(cmd.Args.(*Team))
			case "RemoveTeam": // from world
				w.removeTeam(cmd.Args.(int64))
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
