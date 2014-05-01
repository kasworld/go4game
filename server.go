package go4game

import (
	"fmt"
	"log"
	"net"
	//"reflect"
	//"encoding/json"
	//"flag"
	"github.com/gorilla/websocket"
	"net/http"
	//"runtime"
	//"text/template"
	"time"
)

type GameService struct {
	PacketStat
	ID     int
	CmdCh  chan Cmd
	Worlds map[int]*World

	ListenTo           string
	clientConnectionCh chan net.Conn

	wsClientConnectionCh chan *websocket.Conn
	wsListen             string
}

func (m GameService) String() string {
	return fmt.Sprintf("GameService:%v Worlds:%v", m.ID, len(m.Worlds))
}

func NewGameService(listenTo string, wsListen string) *GameService {
	g := GameService{
		ID:                   <-IdGenCh,
		PacketStat:           *NewStatInfo(),
		CmdCh:                make(chan Cmd, 10),
		Worlds:               make(map[int]*World),
		ListenTo:             listenTo,
		clientConnectionCh:   make(chan net.Conn),
		wsClientConnectionCh: make(chan *websocket.Conn),
		wsListen:             wsListen,
	}
	go g.listenLoop()
	go g.wsServer()
	log.Printf("New %v", g)
	go g.Loop()

	// create default world
	g.addNewWorld()
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

func (g *GameService) findFreeWorld(teamCount int, ct ClientType) *World {
	for _, w := range g.Worlds {
		if w.teamCount(ct) < teamCount {
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
			w := g.findFreeWorld(32, TCPClient)
			w.CmdCh <- Cmd{
				Cmd:  "newTeam",
				Args: conn,
			}
		case conn := <-g.wsClientConnectionCh: // new team
			w := g.findFreeWorld(32, WebSockClient)
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
				s := cmd.Args.(PacketStat)
				g.PacketStat.AddLap(&s)
			case "delWorld":
				g.delWorld(cmd.Args.(*World))
			default:
				log.Printf("unknown cmd %v", cmd)
			}
		case <-timer60Ch:
			// do frame action
		case <-timer1secCh:
			// log.Printf("%v ID:%v goroutine:%v\n%v",
			// 	g, <-IdGenCh, runtime.NumGoroutine(), g.PacketStat)
			g.PacketStat.NewLap()
		}
	}
	log.Printf("quit %v", g)
}

func (g *GameService) listenLoop() {
	listener, err := net.Listen("tcp", g.ListenTo)
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

// web socket server
func (g *GameService) wsServer() {
	http.HandleFunc("/ws", g.wsServe)
	http.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("./www"))))
	err := http.ListenAndServe(g.wsListen, nil)
	if err != nil {
		log.Println("ListenAndServe: ", err)
	}
}

func (g *GameService) wsServe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	g.wsClientConnectionCh <- ws
}
