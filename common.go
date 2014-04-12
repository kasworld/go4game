package go4game

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
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

type StatInfo struct {
	RCount    int
	RLen      int
	WCount    int
	WLen      int
	StartTime time.Time
}

func NewStatInfo() *StatInfo {
	return &StatInfo{
		RCount:    0,
		RLen:      0,
		WCount:    0,
		WLen:      0,
		StartTime: time.Now(),
	}
}

func (d *StatInfo) Reset() {
	d.RCount = 0
	d.RLen = 0
	d.WCount = 0
	d.WLen = 0
	d.StartTime = time.Now()
}

func (d *StatInfo) Add(s *StatInfo) {
	d.RCount += s.RCount
	d.RLen += s.RLen
	d.WCount += s.WCount
	d.WLen += s.WLen
}

func (d *StatInfo) AddR(c int, l int) {
	d.RCount += c
	d.RLen += l
}

func (d *StatInfo) AddW(c int, l int) {
	d.WCount += c
	d.WLen += l
}

func (d *StatInfo) ToString() string {
	dur := time.Now().Sub(d.StartTime)
	return fmt.Sprintf("Stat:rcount:%v/%v rlen:%v/%v wcount:%v/%v wlen:%v/%v",
		d.RCount, float64(d.RCount)/dur.Seconds(),
		d.RLen, float64(d.RLen)/dur.Seconds(),
		d.WCount, float64(d.WCount)/dur.Seconds(),
		d.WLen, float64(d.WLen)/dur.Seconds())
}

type Cmd struct {
	Cmd  string
	Args interface{}
}

type ConnInfo struct {
	Stat    *StatInfo
	CmdCh   chan Cmd
	Conn    net.Conn
	ReadCh  chan interface{}
	WriteCh chan interface{}
}

func NewConnInfo(conn net.Conn) *ConnInfo {
	c := ConnInfo{
		Stat:    NewStatInfo(),
		CmdCh:   make(chan Cmd),
		Conn:    conn,
		ReadCh:  make(chan interface{}, 10),
		WriteCh: make(chan interface{}, 10),
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
			if err.Error() != "EOF" {
				log.Print(err)
			}
			break
		} else {
			c.ReadCh <- packet
			c.Stat.AddR(1, int(n))
		}
	}
}

func (c *ConnInfo) writeLoop() {
	defer c.Conn.Close()
writeloop:
	for {
		select {
		case packet := <-c.WriteCh:
			n, err := writeJson(c.Conn, packet)
			if err != nil {
				if err.Error() != "EOF" {
					log.Print(err)
				}
				break writeloop
			}
			c.Stat.AddW(1, int(n))
		}
	}
}
