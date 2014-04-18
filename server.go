package go4game

import (
	"fmt"
	"log"
	"net"
	//"reflect"
	"runtime"
	"time"
)

type GameService struct {
	StatInfo
	ID    int
	CmdCh chan Cmd
	//Name               string
	Worlds             map[int]*World
	ListenTo           string
	clientConnectionCh chan net.Conn
}

func (m GameService) String() string {
	return fmt.Sprintf("GameService:%v Worlds:%v", m.ID, len(m.Worlds))
}

func NewGameService(listenTo string) *GameService {
	g := GameService{
		ID:       <-IdGenCh,
		StatInfo: *NewStatInfo(),
		CmdCh:    make(chan Cmd, 10),
		Worlds:   make(map[int]*World),
		ListenTo: listenTo,
	}
	g.clientConnectionCh = g.listenLoop()
	//g.addNewWorld()
	log.Printf("New %v", g)
	go g.Loop()
	return &g
}

func (g *GameService) addNewWorld() *World {
	w := NewWorld(g)
	g.Worlds[w.ID] = w
	return w
}

func (g *GameService) delWorld(w *World) {
	delete(g.Worlds, w.ID)
}

func (g *GameService) findFreeWorld(teamCount int) *World {
	for _, w := range g.Worlds {
		if len(w.Teams) < teamCount {
			return w
		}
	}
	return g.addNewWorld()
}

func (g *GameService) Loop() {
	timer1secCh := time.Tick(1 * time.Second)
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
loop:
	for {
		select {
		case conn := <-g.clientConnectionCh: // new team
			w := g.findFreeWorld(32)
			w.CmdCh <- Cmd{
				Cmd:  "newTeam",
				Args: conn,
			}
		case cmd := <-g.CmdCh:
			//log.Println(cmd)
			switch cmd.Cmd {
			case "quit":
				for _, v := range g.Worlds {
					v.CmdCh <- Cmd{Cmd: "quit"}
				}
				break loop
			case "statInfo":
				s := cmd.Args.(StatInfo)
				g.StatInfo.AddLap(&s)
			case "delWorld":
				g.delWorld(cmd.Args.(*World))
			default:
				log.Printf("unknown cmd %v", cmd)
			}
		case <-timer60Ch:
			// do frame action
		case <-timer1secCh:
			tsum := 0
			for _, w := range g.Worlds {
				tsum += len(w.Teams)
			}
			log.Printf("%v teams:%v goroutine:%v\n%v",
				g, tsum, runtime.NumGoroutine(), g.StatInfo)
			g.StatInfo.NewLap()
		}
	}
	log.Printf("quit %v", g)
}

func (g *GameService) listenLoop() chan net.Conn {
	clientConnectionCh := make(chan net.Conn)
	go func(clientConnectionCh chan net.Conn) {
		listener, err := net.Listen("tcp", g.ListenTo)
		if err != nil {
			log.Print(err)
			//os.Exit(1)
		}
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Print(err)
			}
			clientConnectionCh <- conn
		}
	}(clientConnectionCh)
	return clientConnectionCh
}
