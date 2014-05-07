package go4game

import (
	//"encoding/binary"
	//"errors"
	//"fmt"
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
