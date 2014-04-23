package go4game

import (
	//"encoding/binary"
	"encoding/json"
	//"errors"
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

type ConnInfo struct {
	Stat    *PacketStat
	CmdCh   chan Cmd
	PTeam   *Team
	Conn    net.Conn
	ReadCh  chan *GamePacket
	WriteCh chan *GamePacket
	// ReadCh  chan []byte
	// WriteCh chan interface{}
}

func NewConnInfo(t *Team, conn net.Conn) *ConnInfo {
	c := ConnInfo{
		Stat:    NewStatInfo(),
		CmdCh:   make(chan Cmd, 2),
		Conn:    conn,
		ReadCh:  make(chan *GamePacket, 1),
		WriteCh: make(chan *GamePacket, 1),
		// ReadCh:  make(chan []byte, 1),
		// WriteCh: make(chan interface{}, 1),
		PTeam: t,
	}
	go c.readLoop()
	go c.writeLoop()
	return &c
}

func (c *ConnInfo) readLoop() {
	defer c.Conn.Close()
	dec := json.NewDecoder(c.Conn)
	for {
		//buf, err := readI32Packet(c.Conn)
		var v GamePacket
		err := dec.Decode(&v)
		if err != nil {
			c.PTeam.CmdCh <- Cmd{Cmd: "quitRead", Args: err}
			break
		} else {
			c.ReadCh <- &v
			c.Stat.IncR()
		}
	}
	//log.Print("quit read")
}

// func readI32Packet(conn net.Conn) ([]byte, error) {
// 	var hlen uint32
// 	err := binary.Read(conn, binary.LittleEndian, &hlen)
// 	if err != nil {
// 		return nil, err
// 	}

// 	buf := make([]byte, hlen)
// 	n, err := conn.Read(buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if uint32(n) != hlen {
// 		errmsg := fmt.Sprintf("read incomplete %v %v", n, hlen)
// 		return buf, errors.New(errmsg)
// 	}
// 	return buf, nil
// }

// func readJson(conn net.Conn, recvpacket interface{}) (int, error) {
// 	buf, err := readI32Packet(conn)
// 	//log.Printf("%v", string(buf))
// 	if err != nil {
// 		return len(buf), err
// 	}
// 	return len(buf), json.Unmarshal(buf, &recvpacket)
// }

func (c *ConnInfo) writeLoop() {
	defer c.Conn.Close()
	enc := json.NewEncoder(c.Conn)
writeloop:
	for {
		select {
		case cmd := <-c.CmdCh:
			switch cmd.Cmd {
			case "quit":
				break writeloop
			}
		case packet := <-c.WriteCh:
			//n, err := writeJson(c.Conn, packet)
			err := enc.Encode(packet)
			if err != nil {
				c.PTeam.CmdCh <- Cmd{Cmd: "quitWrite", Args: err}
				break writeloop
			}
			c.Stat.IncW()
		}
	}
	//log.Print("quit write")
}

// func writeI32Packet(conn net.Conn, packet []byte) (int, error) {
// 	hlen := len(packet)
// 	err := binary.Write(conn, binary.LittleEndian, uint32(hlen))
// 	if err != nil {
// 		return 0, err
// 	}
// 	n, err := conn.Write(packet)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if n != hlen {
// 		errmsg := fmt.Sprintf("write incomplete %v %v", n, hlen)
// 		return n, errors.New(errmsg)
// 	}
// 	return hlen, nil
// }

// func writeJson(conn net.Conn, sendPacket interface{}) (int, error) {
// 	packet, err := json.Marshal(sendPacket)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return writeI32Packet(conn, packet)
// }
