package go4game

import (
	//"encoding/binary"
	//"errors"
	//"fmt"
	"log"
	"math/rand"
	"runtime"
	"sort"
	"time"
)

type IDList []int64

func (s IDList) Len() int {
	return len(s)
}
func (s IDList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s IDList) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s IDList) findIndex(id int64) int {
	return sort.Search(len(s), func(i int) bool { return s[i] >= id })
}

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

type GoCmd struct {
	Cmd  string
	Args interface{}
	Rsp  chan<- interface{}
}
