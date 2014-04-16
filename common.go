package go4game

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	//"log"
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

type GamePacket struct {
	Cmd string

	Teamname      string
	Teamcolor     []int
	Teamid        int
	TeamStartTime int
}

func readJson(conn net.Conn, recvpacket interface{}) (uint32, error) {
	var hlen uint32
	err := binary.Read(conn, binary.LittleEndian, &hlen)
	if err != nil {
		return hlen, err
	}

	buf := make([]byte, hlen)
	n, err := conn.Read(buf)
	if err != nil {
		return hlen, err
	}
	if uint32(n) != hlen {
		errmsg := fmt.Sprintf("read incomplete %v %v", n, hlen)
		return uint32(n), errors.New(errmsg)
	}

	err = json.Unmarshal(buf, &recvpacket)
	if err != nil {
		return hlen, err
	}
	return hlen, nil
}

func writeJson(conn net.Conn, sendPacket interface{}) (uint32, error) {
	packet, err := json.Marshal(sendPacket)
	if err != nil {
		return 0, err
	}
	hlen := uint32(len(packet))
	err = binary.Write(conn, binary.LittleEndian, hlen)
	if err != nil {
		return 0, err
	}
	n, err := conn.Write(packet)
	if err != nil {
		return 0, err
	}
	if uint32(n) != hlen {
		errmsg := fmt.Sprintf("write incomplete %v %v", n, hlen)
		return uint32(n), errors.New(errmsg)
	}
	return hlen, nil
}

type CountLen struct {
	Count int
	Len   int
}

func (cl *CountLen) Inc(l int) {
	cl.Count += 1
	cl.Len += l
}

func (cl *CountLen) Add(c *CountLen) {
	cl.Count += c.Count
	cl.Len += c.Len
}

func (cl *CountLen) Clear() {
	cl.Count = 0
	cl.Len = 0
}

func (cl CountLen) CalcLap(dur time.Duration) string {
	return fmt.Sprintf("count:%v/%5.1f len:%v/%5.1f",
		cl.Count, float64(cl.Count)/dur.Seconds(),
		cl.Len, float64(cl.Len)/dur.Seconds())
}

type StatInfo struct {
	ReadCL      CountLen
	WriteCL     CountLen
	ReadSum     CountLen
	WriteSum    CountLen
	StartTime   time.Time
	LastLapTime time.Time
}

func (d StatInfo) String() string {
	lapdur := time.Now().Sub(d.LastLapTime)
	dur := time.Now().Sub(d.StartTime)
	return fmt.Sprintf("recv(total:%v lap:%v)\nsend(total:%v lap:%v)",
		d.ReadSum.CalcLap(dur),
		d.ReadCL.CalcLap(lapdur),
		d.WriteSum.CalcLap(dur),
		d.WriteCL.CalcLap(lapdur))
}

func NewStatInfo() *StatInfo {
	return &StatInfo{
		StartTime:   time.Now(),
		LastLapTime: time.Now(),
	}
}

func (d *StatInfo) NewLap() {
	d.ReadCL.Clear()
	d.WriteCL.Clear()
	d.LastLapTime = time.Now()
}

func (d *StatInfo) AddLap(s *StatInfo) {
	d.ReadCL.Add(&s.ReadCL)
	d.WriteCL.Add(&s.WriteCL)
	d.ReadSum.Add(&s.ReadCL)
	d.WriteSum.Add(&s.WriteCL)
}

func (d *StatInfo) IncR(l int) {
	d.ReadCL.Inc(l)
	d.ReadSum.Inc(l)
}

func (d *StatInfo) IncW(l int) {
	d.WriteCL.Inc(l)
	d.WriteSum.Inc(l)
}

type Cmd struct {
	Cmd  string
	Args interface{}
}

type ConnInfo struct {
	Stat    *StatInfo
	CmdCh   chan Cmd
	PTeam   *Team
	Conn    net.Conn
	ReadCh  chan interface{}
	WriteCh chan interface{}
}

func NewConnInfo(t *Team, conn net.Conn) *ConnInfo {
	c := ConnInfo{
		Stat:    NewStatInfo(),
		CmdCh:   make(chan Cmd, 2),
		Conn:    conn,
		ReadCh:  make(chan interface{}, 1),
		WriteCh: make(chan interface{}, 1),
		PTeam:   t,
	}
	go c.readLoop()
	go c.writeLoop()
	return &c
}

func (c *ConnInfo) readLoop() {
	defer c.Conn.Close()
	for {
		var packet GamePacket
		n, err := readJson(c.Conn, &packet)
		if err != nil {
			c.PTeam.CmdCh <- Cmd{Cmd: "quitRead", Args: err}
			break
		} else {
			c.ReadCh <- packet
			c.Stat.IncR(int(n))
		}
	}
	//log.Print("quit read")
}

func (c *ConnInfo) writeLoop() {
	defer c.Conn.Close()
writeloop:
	for {
		select {
		case cmd := <-c.CmdCh:
			switch cmd.Cmd {
			case "quit":
				break writeloop
			}
		case packet := <-c.WriteCh:
			n, err := writeJson(c.Conn, packet)
			if err != nil {
				c.PTeam.CmdCh <- Cmd{Cmd: "quitWrite", Args: err}
				break writeloop
			}
			c.Stat.IncW(int(n))
		}
	}
	//log.Print("quit write")
}
