package go4game

import (
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
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

func (g *GameService) Loop() {
	<-g.CmdCh
	timer1secCh := time.Tick(1 * time.Second)
	timer60Ch := time.Tick(time.Duration(1000/GameConst.FramePerSec) * time.Millisecond)
loop:
	for {
		select {
		case conn := <-g.clientConnectionCh: // new team
			w := g.findFreeWorld(GameConst.MaxTcpClientPerWorld, TCPClient)
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(conn)}
		case conn := <-g.wsClientConnectionCh: // new team
			w := g.findFreeWorld(GameConst.MaxWsClientPerWorld, WebSockClient)
			w.CmdCh <- Cmd{Cmd: "AddTeam", Args: NewTeam(conn)}
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

// web socket server
func (g *GameService) wsServer() {
	http.HandleFunc("/ws", g.wsServe)
	http.HandleFunc("/", g.Stat)
	http.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("./www"))))
	err := http.ListenAndServe(GameConst.WsListen, nil)
	if err != nil {
		log.Println("ListenAndServe: ", err)
	}
}

var TopHtmlTemplate *template.Template
var WorldHtmlTemplate *template.Template

func init() {
	const tindex = `
		<html>
		<head>
		<title>go4game stat</title>
		<meta http-equiv="refresh" content="1">
		</head>
		<body>
		<a href='www/client3d.html' target="_blank">Open 3d client</a>
		</br>
		{{.Disp}}
		</br>
		{{range $id, $s := .Worlds}}
		<a href='?worldid={{$id}}' target="_blank">{{$s}}</a>
		</br>
		{{end}}
		</body>
		</html>
		`
	TopHtmlTemplate = template.Must(template.New("indexpage").Parse(tindex))

	const tworld = `
		<html>
		<head>
		<title>go4game stat</title>
		<meta http-equiv="refresh" content="1">
		</head>
		<body>
		<a href='www/client3d.html' target="_blank">Open 3d client</a>
		</br>
		{{.Disp}}
		</br>
		<table>
		<tr >
			<td>TeamID</td>
			<td>ClientInfo</td>
			<td>ObjCount</td>
			<td>Score</td>
			<td>ActionPoint</td>
			<td>PacketStat</td>
			<td>CollStat</td>
		</tr>
		{{range .Teams}}
		<tr bgcolor="#{{.Color | printf "%x"}}">
			<td><font color="#{{.FontColor | printf "%x"}}">{{.ID}}</font></td>
			<td><font color="#{{.FontColor | printf "%x"}}">{{.ClientInfo}}</font></td>
			<td><font color="#{{.FontColor | printf "%x"}}">{{.Objs}}</font></td>
			<td><font color="#{{.FontColor | printf "%x"}}">{{.Score}}</font></td>
			<td><font color="#{{.FontColor | printf "%x"}}">{{.AP}}</font></td>
			<td><font color="#{{.FontColor | printf "%x"}}">{{.PacketStat}}</font></td>
			<td><font color="#{{.FontColor | printf "%x"}}">{{.CollStat}}</font></td>
		</tr>
		{{end}}
		</table>
		</body>
		</html>
		`
	WorldHtmlTemplate = template.Must(template.New("indexpage").Parse(tworld))
}

func (g *GameService) Stat(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "parse error", 405)
	}
	wid := r.Form.Get("worldid")
	worldid, err := strconv.ParseInt(wid, 0, 64)
	//log.Printf("worldid %v, %v", worldid, err)

	if err != nil {
		ws := make(map[int64]string, len(g.Worlds))
		for id, w := range g.Worlds {
			ws[id] = w.String()
		}
		TopHtmlTemplate.Execute(w, struct {
			Disp   string
			Worlds map[int64]string
		}{
			Disp:   g.String(),
			Worlds: ws,
		})
	} else {
		wi := g.Worlds[worldid].makeWorldInfo()
		WorldHtmlTemplate.Execute(w, wi)
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
