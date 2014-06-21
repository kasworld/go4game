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
		case <-timer60Ch:
			// do frame action
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
func (w *World) NewObjGroup() ObjGroupI {
	og := ObjGroupBase{
		id:       <-go4game.IdGenCh,
		GameObjs: make(map[int64]GameObjI),
		config:   w.config,
	}
	return &og
}
func (w *World) NewSnake() ObjGroupI {
	og := Snake{
		ObjGroupBase: *w.NewObjGroup().(*ObjGroupBase),
	}
	og.AddInitMembers()
	return &og
}
func (w *World) NewStageWalls() ObjGroupI {
	og := StageWalls{
		ObjGroupBase: *w.NewObjGroup().(*ObjGroupBase),
	}
	og.AddInitMembers()
	return &og
}
func (w *World) NewStagePlums() ObjGroupI {
	og := StagePlums{
		ObjGroupBase: *w.NewObjGroup().(*ObjGroupBase),
	}
	og.AddInitMembers()
	return &og
}
func (w *World) NewStageApples() ObjGroupI {
	og := StageApples{
		ObjGroupBase: *w.NewObjGroup().(*ObjGroupBase),
	}
	og.AddInitMembers()
	return &og
}
