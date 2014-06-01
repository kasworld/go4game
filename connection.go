package go4game

import (
	//"encoding/binary"
	"encoding/gob"
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
type MakeAI func() AIActor

// client conn type
type ClientType int

const (
	_ ClientType = iota
	TCPClient
	WebSockClient
	AIClient
)

type ConnInfo struct {
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

func NewAIConnInfo(aiconn AIActor) *ConnInfo {
	c := ConnInfo{
		ReadCh:     make(chan *GamePacket, 1),
		WriteCh:    make(chan *GamePacket, 1),
		AiConn:     aiconn,
		clientType: AIClient,
	}
	go c.aiLoop()
	return &c
}

func (c *ConnInfo) aiLoop() {
	defer func() {
		close(c.ReadCh)
	}()
	c.ReadCh <- &GamePacket{Cmd: ReqFrameInfo}
loop:
	for packet := range c.WriteCh { // get rsp from server
		switch packet.Cmd {
		case RspFrameInfo:
			c.ReadCh <- c.AiConn.MakeAction(packet)
		default:
			log.Printf("unknown packet %v", packet.Cmd)
			break loop
		}
	}
}

type IDecoder interface {
	Decode(v interface{}) error
}
type IEncoder interface {
	Encode(v interface{}) error
}

func NewTcpConnInfo(conn net.Conn) *ConnInfo {
	c := ConnInfo{
		Conn:       conn,
		ReadCh:     make(chan *GamePacket, 1),
		WriteCh:    make(chan *GamePacket, 1),
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
	}()
	var dec IDecoder
	if GameConst.TcpClientEncode == "gob" {
		dec = gob.NewDecoder(c.Conn)
	} else if GameConst.TcpClientEncode == "json" {
		dec = json.NewDecoder(c.Conn)
	} else {
		log.Fatal("unknown tcp client encode %v", GameConst.TcpClientEncode)
	}

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
	}()
	var enc IEncoder
	if GameConst.TcpClientEncode == "gob" {
		enc = gob.NewEncoder(c.Conn)
	} else if GameConst.TcpClientEncode == "json" {
		enc = json.NewEncoder(c.Conn)
	} else {
		log.Fatal("unknown tcp client encode %v", GameConst.TcpClientEncode)
	}
loop:
	for packet := range c.WriteCh {
		err := enc.Encode(packet)
		if err != nil {
			break loop
		}
	}
}

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 0xffff              // Maximum message size allowed from peer.
)

func NewWsConnInfo(conn *websocket.Conn) *ConnInfo {
	c := ConnInfo{
		WsConn:     conn,
		ReadCh:     make(chan *GamePacket, 1),
		WriteCh:    make(chan *GamePacket, 1),
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
