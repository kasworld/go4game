package go4game

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"time"
)

type GameService struct {
	StatInfo
	ID                 int
	CmdCh              chan Cmd
	Name               string
	Worlds             map[int]World
	ListenTo           string
	clientConnectionCh chan net.Conn
}

func (m *GameService) ToString() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.Name, len(m.Worlds))
}

func NewGameService(listenTo string) *GameService {
	g := GameService{
		ID:       <-IdGenCh,
		StatInfo: *NewStatInfo(),
		CmdCh:    make(chan Cmd),
		Worlds:   make(map[int]World),
		ListenTo: listenTo,
	}
	g.clientConnectionCh = g.listenLoop()
	g.addNewWorld()
	log.Printf("new %v", g.ToString())
	go g.Loop()
	return &g
}

func (g *GameService) addNewWorld() {
	w := *NewWorld(g)
	g.Worlds[w.ID] = w
}

func (g *GameService) delWorld(w *World) {
	delete(g.Worlds, w.ID)
}

func (g *GameService) Loop() {
	timer1secCh := time.Tick(1 * time.Second)
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
	for {
		select {
		case <-timer60Ch:
			// do frame action
		case cmd := <-g.CmdCh:
			switch cmd.Cmd {
			case "quit":
				for _, v := range g.Worlds {
					v.CmdCh <- Cmd{Cmd: "quit"}
				}
			case "statInfo":
				g.StatInfo.Add(cmd.Args.(*StatInfo))
			default:
				log.Printf("unknown cmd %v", cmd)
			}
		case conn := <-g.clientConnectionCh: // new team
			for _, v := range g.Worlds {
				v.CmdCh <- Cmd{
					Cmd:  "newTeam",
					Args: conn,
				}
				break
			}
		case <-timer1secCh:
			log.Printf("service:%v\n", g.StatInfo.ToString())
			g.StatInfo.Reset()
		}
	}
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
