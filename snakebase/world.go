package snakebase

import (
	//"encoding/json"
	"fmt"
	"github.com/kasworld/go4game"
	"log"
	//"os"
	//"runtime"
	"time"
)

type World struct {
	id        int64
	ObjGroups map[int64]ObjGroupI
	config    *SnakeConfig
	cmdCh     chan go4game.GoCmd
	pService  *SnakeService
}

func NewWorld(s *SnakeService) *World {
	w := World{
		id:        <-go4game.IdGenCh,
		ObjGroups: make(map[int64]ObjGroupI),
		cmdCh:     make(chan go4game.GoCmd, 1),
		pService:  s,
		config:    s.config,
	}
	w.AddObjGroup(NewSnake(&w))
	w.AddObjGroup(NewStageWalls(&w))
	w.AddObjGroup(NewStageApples(&w))
	w.AddObjGroup(NewStagePlums(&w))
	go w.Loop()
	return &w
}

func (w *World) ID() int64 {
	return w.id
}

func (w *World) SendGoCmd(Cmd string, Args interface{}, Rsp chan<- interface{}) {
	w.cmdCh <- go4game.GoCmd{
		Cmd:  Cmd,
		Args: Args,
		Rsp:  Rsp,
	}
}
func (w World) String() string {
	return fmt.Sprintf("World%v ", w.ID)
}
func (w *World) Do1Frame(ftime time.Time) bool {
	for _, og := range w.ObjGroups {
		go og.StartFrameAction(w, ftime)
	}
	for _, og := range w.ObjGroups {
		og.FrameActionResult()
	}
	return true
}
func (w *World) Loop() {
	timer1secCh := time.Tick(1 * time.Second)
	fps := 60
	timer60Ch := time.Tick(time.Duration(1000/fps) * time.Millisecond)
loop:
	for {
		select {
		case cmd := <-w.cmdCh:
			//log.Println(cmd)
			switch cmd.Cmd {
			default:
				log.Printf("unknown cmd %v", cmd)
			case "quit":
				break loop
			}
		case ftime := <-timer60Ch:
			ok := w.Do1Frame(ftime)
			if !ok {
				break loop
			}
		case <-timer1secCh:
		}
	}
}
func (w *World) AddObjGroup(og ObjGroupI) {
	w.ObjGroups[og.ID()] = og
}
func (w *World) RemoveObjGroup(id int64) {
	delete(w.ObjGroups, id)
}
func (w *World) CollideList(o GameObjI) []GameObjI {
	rtn := make([]GameObjI, 0)
	return rtn
}

func test_WorldI() {
	var w WorldI = &World{}
	_ = w
}
