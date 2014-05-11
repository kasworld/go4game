package go4game

import (
	//"encoding/binary"
	"encoding/json"
	//"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"reflect"
	"time"
)

type AIActor interface {
	MakeAction(*GamePacket) *GamePacket
}

type ConnInfo struct {
	PTeam      *Team
	ReadCh     chan *GamePacket
	WriteCh    chan *GamePacket
	clientType ClientType
	Conn       net.Conn
	WsConn     *websocket.Conn
	AiConn     AIActor
}

func (c ConnInfo) String() string {
	if c.AiConn == nil {
		return fmt.Sprintf("Client%v", c.clientType)
	} else {
		return fmt.Sprintf("AI%v", reflect.TypeOf(c.AiConn))
	}
}

func NewAIConnInfo(t *Team, aiconn AIActor) *ConnInfo {
	c := ConnInfo{
		ReadCh:     make(chan *GamePacket, 1),
		WriteCh:    make(chan *GamePacket, 1),
		PTeam:      t,
		AiConn:     aiconn,
		clientType: AIClient,
	}
	//aiconn.pteam = t
	go c.aiLoop()
	return &c
}

func (c *ConnInfo) aiLoop() {
	defer func() {
		//log.Printf("aiLoop end team:%v", c.PTeam.ID)
		close(c.ReadCh)
	}()
	c.ReadCh <- &GamePacket{
		Cmd: ReqFrameInfo,
	}
loop:
	for {
		select {
		case packet, ok := <-c.WriteCh: // get rsp from server
			if !ok {
				break loop
			}
			switch packet.Cmd {
			case RspFrameInfo:
				c.ReadCh <- c.AiConn.MakeAction(packet)
			default:
				log.Printf("unknown packet %v", packet.Cmd)
				break loop
			}
		}
	}
}

func NewTcpConnInfo(t *Team, conn net.Conn) *ConnInfo {
	c := ConnInfo{
		Conn:       conn,
		ReadCh:     make(chan *GamePacket, 1),
		WriteCh:    make(chan *GamePacket, 1),
		PTeam:      t,
		clientType: TCPClient,
	}
	go c.tcpReadLoop()
	go c.tcpWriteLoop()
	return &c
}

func (c *ConnInfo) tcpReadLoop() {
	defer func() {
		c.Conn.Close()
		close(c.ReadCh)
		//log.Printf("tcpReadLoop end team:%v", c.PTeam.ID)
	}()
	dec := json.NewDecoder(c.Conn)
	for {
		var v GamePacket
		err := dec.Decode(&v)
		if err != nil {
			break
		}
		c.ReadCh <- &v
	}
}

func (c *ConnInfo) tcpWriteLoop() {
	defer func() {
		c.Conn.Close()
		//log.Printf("tcpWriteLoop end team:%v", c.PTeam.ID)
	}()
	enc := json.NewEncoder(c.Conn)
loop:
	for {
		select {
		case packet, ok := <-c.WriteCh:
			if !ok {
				break loop
			}
			err := enc.Encode(packet)
			if err != nil {
				break loop
			}
		}
	}
}

func NewWsConnInfo(t *Team, conn *websocket.Conn) *ConnInfo {
	c := ConnInfo{
		WsConn:     conn,
		ReadCh:     make(chan *GamePacket, 1),
		WriteCh:    make(chan *GamePacket, 1),
		PTeam:      t,
		clientType: WebSockClient,
	}
	go c.wsReadLoop()
	go c.wsWriteLoop()
	return &c
}

func (c *ConnInfo) wsReadLoop() {
	defer func() {
		c.WsConn.Close()
		close(c.ReadCh)
		//log.Printf("wsReadLoop end team:%v", c.PTeam.ID)
	}()
	c.WsConn.SetReadLimit(maxMessageSize)
	c.WsConn.SetReadDeadline(time.Now().Add(pongWait))
	c.WsConn.SetPongHandler(func(string) error {
		c.WsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var v GamePacket
		err := c.WsConn.ReadJSON(&v)
		if err != nil {
			break
		}
		c.ReadCh <- &v
	}
}

func (c *ConnInfo) write(mt int, payload []byte) error {
	c.WsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.WsConn.WriteMessage(mt, payload)
}

func (c *ConnInfo) wsWriteLoop() {
	timerPing := time.Tick(pingPeriod)
	defer func() {
		c.WsConn.Close()
		//log.Printf("wsWriteLoop end team:%v", c.PTeam.ID)
	}()
	for {
		select {
		case packet, ok := <-c.WriteCh:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			message, err := json.Marshal(&packet)
			if err != nil {
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-timerPing:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
