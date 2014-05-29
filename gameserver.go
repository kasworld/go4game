package go4game

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	// "net/http"
	"runtime"
	// "strconv"
	"time"
)

type GameService struct {
	ID     int64
	CmdCh  chan Cmd
	Worlds map[int64]*World

	clientConnectionCh   chan net.Conn
	wsClientConnectionCh chan *websocket.Conn
}

func (m GameService) String() string {
	return fmt.Sprintf("GameService%v Worlds:%v goroutine:%v IDs:%v",
		m.ID, len(m.Worlds), runtime.NumGoroutine(), <-IdGenCh)
}

func NewGameService() *GameService {
	g := GameService{
		ID:                   <-IdGenCh,
		CmdCh:                make(chan Cmd, 10),
		Worlds:               make(map[int64]*World),
		clientConnectionCh:   make(chan net.Conn),
		wsClientConnectionCh: make(chan *websocket.Conn),
	}
	go g.listenLoop()
	go g.wsServer()

	log.Printf("New %v\n%+v", g, GameConst)

	// create default world
	for i := 0; i < GameConst.StartWorldCount; i++ {
		g.addNewWorld()
	}
	return &g
}

func (g *GameService) addNewWorld() *World {
	w := NewWorld(g)
	g.Worlds[w.ID] = w
	go w.Loop()
	w.addAITeams(GameConst.AINames, GameConst.AICountPerWorld)
	return w
}

func (g *GameService) delWorld(w *World) {
	delete(g.Worlds, w.ID)
}

func (g *GameService) findFreeWorld(teamCount int, ct ClientType) *World {
	for _, w := range g.Worlds {
		if w.teamCount(ct) < teamCount {
			return w
		}
	}
	return g.addNewWorld()
}

func (g *GameService) MoveTeam(w1id, w2id int64, tid int64) bool {
	if w1id == w2id {
		return false
	}
	w1 := g.Worlds[w1id]
	w2 := g.Worlds[w2id]
	if w1 == nil || w2 == nil {
		return false
	}
	t := w1.Teams[tid]
	if t == nil {
		return false
	}
	//log.Printf("remove team%v from world%v ", tid, w1id)
	rsp := make(chan interface{})
	w1.CmdCh <- Cmd{Cmd: "RemoveTeam", Args: tid, Rsp: rsp}
	<-rsp
	//log.Printf("add team%v to world%v", tid, w2id)
	w2.CmdCh <- Cmd{Cmd: "AddTeam", Args: t, Rsp: rsp}
	<-rsp
	//log.Printf("end team%v from world%v to world%v", tid, w1id, w2id)
	return true
}

func (g *GameService) MoveTeamRandom() {
	var w1id, w2id, tid int64
	for len(g.Worlds) < 2 {
		g.addNewWorld()
	}
	for id, _ := range g.Worlds {
		if w1id == 0 {
			w1id = id
		} else {
			w2id = id
			break
		}
	}
	if len(g.Worlds[w1id].Teams) > len(g.Worlds[w2id].Teams) {
		for i, t := range g.Worlds[w1id].Teams {
			if t.ClientConnInfo.clientType != AIClient {
				tid = i
				g.MoveTeam(w1id, w2id, tid)
				break
			}
		}
	} else {
		for i, t := range g.Worlds[w2id].Teams {
			if t.ClientConnInfo.clientType != AIClient {
				tid = i
				g.MoveTeam(w2id, w1id, tid)
				break
			}
		}
	}
	if tid == 0 {
		log.Printf("not found %v %v %v", w1id, w2id, tid)
	}
}

func (g *GameService) Loop() {
	<-g.CmdCh
	timer1secCh := time.Tick(1 * time.Second)
	timer60Ch := time.Tick(time.Duration(1000/GameConst.FramePerSec) * time.Millisecond)
loop:
	for {
		select {
		case conn := <-g.clientConnectionCh: // new team
			w := g.findFreeWorld(GameConst.MaxTcpClientPerWorld, TCPClient)
			rsp := make(chan interface{})
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(conn), Rsp: rsp}
			<-rsp
		case conn := <-g.wsClientConnectionCh: // new team
			w := g.findFreeWorld(GameConst.MaxWsClientPerWorld, WebSockClient)
			rsp := make(chan interface{})
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(conn), Rsp: rsp}
			<-rsp
		case cmd := <-g.CmdCh:
			//log.Println(cmd)
			switch cmd.Cmd {
			case "quit":
				for _, v := range g.Worlds {
					v.CmdCh <- Cmd{Cmd: "quit"}
				}
				break loop
			case "delWorld":
				g.delWorld(cmd.Args.(*World))
			default:
				log.Printf("unknown cmd %v", cmd)
			}
		case <-timer60Ch:
			// do frame action
		case <-timer1secCh:
			//g.MoveTeamRandom()
		}
	}
	log.Printf("quit %v", g)
}

func (g *GameService) listenLoop() {
	listener, err := net.Listen("tcp", GameConst.TcpListen)
	if err != nil {
		log.Print(err)
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}
		g.clientConnectionCh <- conn
	}
}
