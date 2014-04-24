package go4game

import (
	//"encoding/binary"
	"encoding/json"
	//"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net"
	"runtime"
	"time"
)

var IdGenCh chan int

func init() {
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())
	IdGenCh = make(chan int)
	go func() {
		i := 0
		for {
			i++
			IdGenCh <- i
		}
	}()
}

type CountStat struct {
	Count int
}

func (cl *CountStat) Inc() {
	cl.Count += 1
}

func (cl *CountStat) Add(c *CountStat) {
	cl.Count += c.Count
}

func (cl *CountStat) Clear() {
	cl.Count = 0
}

func (cl CountStat) CalcLap(dur time.Duration) string {
	return fmt.Sprintf("[%v/%5.1f/s]",
		cl.Count, float64(cl.Count)/dur.Seconds())
}

type PacketStat struct {
	ReadCL      CountStat
	WriteCL     CountStat
	ReadSum     CountStat
	WriteSum    CountStat
	StartTime   time.Time
	LastLapTime time.Time
}

func (d PacketStat) String() string {
	lapdur := time.Now().Sub(d.LastLapTime)
	dur := time.Now().Sub(d.StartTime)
	return fmt.Sprintf("recv(total:%v lap:%v)\nsend(total:%v lap:%v)",
		d.ReadSum.CalcLap(dur),
		d.ReadCL.CalcLap(lapdur),
		d.WriteSum.CalcLap(dur),
		d.WriteCL.CalcLap(lapdur))
}

func NewStatInfo() *PacketStat {
	return &PacketStat{
		StartTime:   time.Now(),
		LastLapTime: time.Now(),
	}
}

func (d *PacketStat) NewLap() {
	d.ReadCL.Clear()
	d.WriteCL.Clear()
	d.LastLapTime = time.Now()
}

func (d *PacketStat) AddLap(s *PacketStat) {
	d.ReadCL.Add(&s.ReadCL)
	d.WriteCL.Add(&s.WriteCL)
	d.ReadSum.Add(&s.ReadCL)
	d.WriteSum.Add(&s.WriteCL)
}

func (d *PacketStat) IncR() {
	d.ReadCL.Inc()
	d.ReadSum.Inc()
}

func (d *PacketStat) IncW() {
	d.WriteCL.Inc()
	d.WriteSum.Inc()
}

type Cmd struct {
	Cmd  string
	Args interface{}
}

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 0xffff              // Maximum message size allowed from peer.
)

type ConnInfo struct {
	Stat    *PacketStat
	PTeam   *Team
	Conn    net.Conn
	WsConn  *websocket.Conn
	ReadCh  chan *GamePacket
	WriteCh chan *GamePacket
}

func NewConnInfo(t *Team, conn net.Conn) *ConnInfo {
	c := ConnInfo{
		Stat:    NewStatInfo(),
		Conn:    conn,
		ReadCh:  make(chan *GamePacket, 1),
		WriteCh: make(chan *GamePacket, 1),
		PTeam:   t,
	}
	go c.readLoop()
	go c.writeLoop()
	return &c
}

func (c *ConnInfo) readLoop() {
	defer func() {
		c.Conn.Close()
		close(c.ReadCh)
	}()
	dec := json.NewDecoder(c.Conn)
	for {
		var v GamePacket
		err := dec.Decode(&v)
		if err != nil {
			break
		}
		c.ReadCh <- &v
		c.Stat.IncR()
	}
}

func (c *ConnInfo) writeLoop() {
	defer func() {
		c.Conn.Close()
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
			c.Stat.IncW()
		}
	}
}

func NewWsConnInfo(t *Team, conn *websocket.Conn) *ConnInfo {
	c := ConnInfo{
		Stat:    NewStatInfo(),
		WsConn:  conn,
		ReadCh:  make(chan *GamePacket, 1),
		WriteCh: make(chan *GamePacket, 1),
		PTeam:   t,
	}
	go c.wsReadLoop()
	go c.wsWriteLoop()
	return &c
}

func (c *ConnInfo) wsReadLoop() {
	defer func() {
		c.WsConn.Close()
		close(c.ReadCh)
		//log.Println("quit wsReadLoop")
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
		c.Stat.IncR()
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
		//log.Println("quit wsWriteLoop")
	}()
	for {
		select {
		case packet, ok := <-c.WriteCh:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			message, err := json.Marshal(&packet)
			log.Println(message)
			if err != nil {
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
			c.Stat.IncW()
		case <-timerPing:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
