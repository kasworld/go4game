package go4game

import (
	//"encoding/binary"
	//"errors"
	//"fmt"
	"math/rand"
	"runtime"
	"time"
)

type IDList []int64

var IdGenCh chan int64

func init() {
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
}
