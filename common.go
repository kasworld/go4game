package go4game

import (
	//"encoding/binary"
	//"errors"
	//"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"
)

type IDList []int64

var IdGenCh chan int64

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetPrefix("go4game ")
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())
	IdGenCh = make(chan int64)
	go func() {
		var i int64
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
	Rsp  chan<- interface{}
}

type ActionPoint struct {
	point int
	as    [ActionEnd]ActionStat
}

func NewActionPoint() *ActionPoint {
	r := ActionPoint{}
	for i := ActionAccel; i < ActionEnd; i++ {
		r.as[i] = *NewActionStat()
	}
	return &r
}

func (ap *ActionPoint) Add(val int) {
	ap.point += val
}

func (ap *ActionPoint) Use(apt ClientActionType, count int) bool {
	if ap.CanUse(apt, count) {
		ap.point -= GameConst.AP[apt]
		ap.as[apt].Inc()
		return true
	}
	return false
}

func (ap *ActionPoint) CanUse(apt ClientActionType, count int) bool {
	return ap.point >= GameConst.AP[apt]*count
}
