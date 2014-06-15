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
	//"reflect"
	"time"
)

type AIActor interface {
	MakeAction(*RspGamePacket) *ReqGamePacket
	String() string
}

// client conn type
type ConnType int

const (
	_ ConnType = iota
	TCPConn
	WebSockConn
	AIConn
)

type ConnInfo struct {
	ReadCh   chan *ReqGamePacket
	WriteCh  chan *RspGamePacket
	connType ConnType
	tcpConn  net.Conn
	wsConn   *websocket.Conn
	aiConn   AIActor
}

func (c ConnInfo) String() string {
	if c.aiConn == nil {
		return fmt.Sprintf("Client%v", c.connType)
	} else {
		return c.aiConn.String() // fmt.Sprintf("AI%v", reflect.TypeOf(c.aiConn))
	}
}

func NewAIConnInfo(aiconn AIActor) *ConnInfo {
	c := ConnInfo{
		ReadCh:   make(chan *ReqGamePacket, 1),
		WriteCh:  make(chan *RspGamePacket, 1),
		aiConn:   aiconn,
		connType: AIConn,
	}
	go c.aiLoop()
	return &c
}

func (c *ConnInfo) aiLoop() {
	defer func() {
		close(c.ReadCh)
	}()
	c.ReadCh <- &ReqGamePacket{Cmd: ReqNearInfo}
loop:
	for packet := range c.WriteCh { // get rsp from server
		switch packet.Cmd {
		case RspNearInfo:
			c.ReadCh <- c.aiConn.MakeAction(packet)
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
		tcpConn:  conn,
		ReadCh:   make(chan *ReqGamePacket, 1),
		WriteCh:  make(chan *RspGamePacket, 1),
		connType: TCPConn,
	}
	go c.tcpReadLoop()
	go c.tcpWriteLoop()
	return &c
}

func (c *ConnInfo) tcpReadLoop() {
	defer func() {
		c.tcpConn.Close()
		close(c.ReadCh)
	}()
	var dec IDecoder
	if GameConst.TcpClientEncode == "gob" {
		dec = gob.NewDecoder(c.tcpConn)
	} else if GameConst.TcpClientEncode == "json" {
		dec = json.NewDecoder(c.tcpConn)
	} else {
		log.Fatal("unknown tcp client encode %v", GameConst.TcpClientEncode)
	}

	for {
		var v ReqGamePacket
		err := dec.Decode(&v)
		if err != nil {
			break
		}
		c.ReadCh <- &v
	}
}

func (c *ConnInfo) tcpWriteLoop() {
	defer func() {
		c.tcpConn.Close()
	}()
	var enc IEncoder
	if GameConst.TcpClientEncode == "gob" {
		enc = gob.NewEncoder(c.tcpConn)
	} else if GameConst.TcpClientEncode == "json" {
		enc = json.NewEncoder(c.tcpConn)
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
		wsConn:   conn,
		ReadCh:   make(chan *ReqGamePacket, 1),
		WriteCh:  make(chan *RspGamePacket, 1),
		connType: WebSockConn,
	}
	go c.wsReadLoop()
	go c.wsWriteLoop()
	return &c
}

func (c *ConnInfo) wsReadLoop() {
	defer func() {
		c.wsConn.Close()
		close(c.ReadCh)
	}()
	c.wsConn.SetReadLimit(maxMessageSize)
	c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	c.wsConn.SetPongHandler(func(string) error {
		c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var v ReqGamePacket
		err := c.wsConn.ReadJSON(&v)
		if err != nil {
			break
		}
		c.ReadCh <- &v
	}
}

func (c *ConnInfo) write(mt int, payload []byte) error {
	c.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.wsConn.WriteMessage(mt, payload)
}

func (c *ConnInfo) wsWriteLoop() {
	timerPing := time.Tick(pingPeriod)
	defer func() {
		c.wsConn.Close()
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
