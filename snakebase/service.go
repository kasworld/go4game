package snakebase

import (
	//"encoding/json"
	"fmt"
	"github.com/kasworld/go4game"
	"log"
	//"os"
	"runtime"
	"time"
)

type SnakeService struct {
	id     int64
	Worlds map[int64]WorldI
	config *SnakeConfig
	cmdCh  chan go4game.GoCmd
}

func (s *SnakeService) ID() int64 {
	return s.id
}

func (s *SnakeService) SendGoCmd(Cmd string, Args interface{}, Rsp chan<- interface{}) {
	s.cmdCh <- go4game.GoCmd{
		Cmd:  Cmd,
		Args: Args,
		Rsp:  Rsp,
	}
}
func (s SnakeService) String() string {
	return fmt.Sprintf("SnakeService%v Worlds:%v goroutine:%v IDs:%v",
		s.ID, len(s.Worlds), runtime.NumGoroutine(), <-go4game.IdGenCh)
}
func (s *SnakeService) Loop() {
	fps := 60
	timer60Ch := time.Tick(time.Duration(1000/fps) * time.Millisecond)
	timer1secCh := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case cmd := <-s.cmdCh:
			//log.Println(cmd)
			switch cmd.Cmd {
			default:
				log.Printf("unknown cmd %v", cmd)
			case "quit":
				break loop
			}
		case <-timer60Ch:
			// do frame action
		case <-timer1secCh:
		}
	}
}
func (s *SnakeService) NewWorld() WorldI {
	w := World{
		id:        <-go4game.IdGenCh,
		ObjGroups: make(map[int64]ObjGroupI),
		cmdCh:     make(chan go4game.GoCmd, 1),
		pService:  s,
		config:    s.config,
	}
	w.AddObjGroup(w.NewSnake())
	w.AddObjGroup(w.NewStageWalls())
	w.AddObjGroup(w.NewStageApples())
	w.AddObjGroup(w.NewStagePlums())
	return &w
}

func (s *SnakeService) AddWorld(w WorldI) {
	ww := w.(*World)
	s.Worlds[ww.ID()] = w
}
func (s *SnakeService) RemoveWorld(id int64) {
	delete(s.Worlds, id)
}
