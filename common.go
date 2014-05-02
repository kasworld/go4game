package go4game

import (
	//"encoding/binary"
	//"errors"
	"fmt"
	"math/rand"
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

func (t *Team) asyncTemplate() <-chan bool {
	chRtn := make(chan bool)
	go func() {
		defer func() { chRtn <- true }()

	}()
	return chRtn
}

type Cmd struct {
	Cmd  string
	Args interface{}
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
	return fmt.Sprintf("recv(total:%v lap:%v) send(total:%v lap:%v)",
		d.ReadSum.CalcLap(dur),
		d.ReadCL.CalcLap(lapdur),
		d.WriteSum.CalcLap(dur),
		d.WriteCL.CalcLap(lapdur))
}

func NewPacketStatInfo() *PacketStat {
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

type LapSumLimit struct {
	Lap         int
	Sum         int
	StartTime   time.Time
	LastLapTime time.Time
	SumLimit    float64
	LapLimit    float64
}

func NewLapSumLimit(sumlimit float64, laplimit float64) *LapSumLimit {
	return &LapSumLimit{
		StartTime:   time.Now(),
		LastLapTime: time.Now(),
		SumLimit:    sumlimit,
		LapLimit:    laplimit,
	}
}

func (d *LapSumLimit) getLapRate() float64 {
	lapdur := time.Now().Sub(d.LastLapTime)
	return float64(d.Lap) / lapdur.Seconds()
}
func (d *LapSumLimit) getSumRate() float64 {
	dur := time.Now().Sub(d.StartTime)
	return float64(d.Sum) / dur.Seconds()
}

func (d LapSumLimit) String() string {
	return fmt.Sprintf("(sum:[%v/%5.1f/s:%v] lap:[%v/%5.1f/s:%v])",
		d.Sum, d.getSumRate(), d.SumLimit,
		d.Lap, d.getLapRate(), d.LapLimit)
}

func (d *LapSumLimit) NewLap() {
	d.Lap = 0
	d.LastLapTime = time.Now()
}

func (d *LapSumLimit) Inc() bool {
	if d.CanInc() {
		d.Lap++
		d.Sum++
		return true
	}
	return false
}

func (d *LapSumLimit) CanInc() bool {
	return d.SumLimit > d.getSumRate() && d.LapLimit > d.getLapRate()
}

type ActStat struct {
	Accel  LapSumLimit
	Bullet LapSumLimit
}

func NewActStat() *ActStat {
	return &ActStat{
		Accel:  *NewLapSumLimit(10, 20),
		Bullet: *NewLapSumLimit(10, 20),
	}
}
func (d ActStat) String() string {
	return fmt.Sprintf("Accel:%v Bullet:%v", d.Accel, d.Bullet)
}

func (d *ActStat) NewLap() {
	d.Accel.NewLap()
	d.Bullet.NewLap()
}
