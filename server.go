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

func (m GameService) String() string {
	return fmt.Sprintf("%v %v %v %v", reflect.TypeOf(m), m.ID, m.Name, len(m.Worlds))
}

func NewGameService(listenTo string) *GameService {
	g := GameService{
		ID:       <-IdGenCh,
		StatInfo: *NewStatInfo(),
		CmdCh:    make(chan Cmd, 10),
		Worlds:   make(map[int]World),
		ListenTo: listenTo,
	}
	g.clientConnectionCh = g.listenLoop()
	g.addNewWorld()
	log.Printf("New %v", g)
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
loop:
	for {
		select {
		case conn := <-g.clientConnectionCh: // new team
			for _, v := range g.Worlds {
				v.CmdCh <- Cmd{
					Cmd:  "newTeam",
					Args: conn,
				}
				break
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
			default:
				log.Printf("unknown cmd %v", cmd)
			}
		case <-timer60Ch:
			// do frame action
		case <-timer1secCh:
			log.Printf("service:%v\n", g.StatInfo.String())
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
