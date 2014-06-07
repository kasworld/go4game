package go4game

import (
	// "net/http"
	// "strconv"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	//"math/rand"
	"net"
	"runtime"
	"sort"
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

func (g *GameService) worldIDList() IDList {
	rtn := make(IDList, 0, len(g.Worlds))
	for id, _ := range g.Worlds {
		rtn = append(rtn, id)
	}
	sort.Sort(rtn)
	return rtn
}

func (g *GameService) addNewWorld() *World {
	w := NewWorld(g)
	g.Worlds[w.ID] = w
	go w.Loop()
	w.addAITeams(GameConst.AINames, GameConst.AICountPerWorld)
	w.addTerrainTeam()
	return w
}

func (g *GameService) delWorld(w *World) {
	delete(g.Worlds, w.ID)
}

func (g *GameService) findFreeWorld(TeamCountByConn int, ct ClientType) *World {
	for _, w := range g.Worlds {
		if w.TeamCountByConn(ct) < TeamCountByConn {
			return w
		}
	}
	return g.addNewWorld()
}

func (g *GameService) nextWorld(wid int64) *World {
	for len(g.Worlds) < 2 {
		return nil
	}
	worldids := g.worldIDList()
	pos := worldids.findIndex(wid)
	if pos < len(worldids) && worldids[pos] == wid { // find and normal
		wid2 := worldids[(pos+1)%len(worldids)]
		return g.Worlds[wid2]
	} else {
		log.Printf("invalid worldid %v", wid)
		return nil
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
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(conn, TeamTypePlayer), Rsp: rsp}
			<-rsp
		case conn := <-g.wsClientConnectionCh: // new team
			w := g.findFreeWorld(GameConst.MaxWsClientPerWorld, WebSockClient)
			rsp := make(chan interface{})
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(conn, TeamTypeObserver), Rsp: rsp}
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
