package snakebase

import (
	"fmt"
	"github.com/kasworld/go4game"
	"log"
	"runtime"
	"time"
)

type SnakeConfig struct {
}

func (config *SnakeConfig) Validate() {
}
func (config *SnakeConfig) Save(filename string) bool {
	return true
}
func (config *SnakeConfig) Load(filename string) bool {
	return true
}
func (config *SnakeConfig) SaveLoad(filename string) {
}
func (config *SnakeConfig) NewService() ServiceI {
	g := SnakeService{
		ID:     <-go4game.IdGenCh,
		cmdCh:  make(chan go4game.GoCmd, 10),
		Worlds: make(map[int64]WorldI),
		config: config,
	}
	return &g
}

var GameConst = SnakeConfig{}

type SnakeService struct {
	ID     int64
	Worlds map[int64]WorldI
	config *SnakeConfig
	cmdCh  chan go4game.GoCmd
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
	timer1secCh := time.Tick(1 * time.Second)
	fps := 60
	timer60Ch := time.Tick(time.Duration(1000/fps) * time.Millisecond)
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
	w := SnakeWorld{
		ID:        <-go4game.IdGenCh,
		ObjGroups: make(map[int64]ObjGroupI),
		cmdCh:     make(chan go4game.GoCmd, 1),
		pService:  s,
		config:    s.config,
	}
	return &w
}

type SnakeWorld struct {
	ID        int64
	ObjGroups map[int64]ObjGroupI
	config    *SnakeConfig
	cmdCh     chan go4game.GoCmd
	pService  *SnakeService
}

func (w *SnakeWorld) SendGoCmd(Cmd string, Args interface{}, Rsp chan<- interface{}) {
	w.cmdCh <- go4game.GoCmd{
		Cmd:  Cmd,
		Args: Args,
		Rsp:  Rsp,
	}
}
func (w SnakeWorld) String() string {
	return fmt.Sprintf("SnakeWorld%v ", w.ID)
}
func (w *SnakeWorld) Loop() {
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
func (w *SnakeWorld) AddObjGroup(og ObjGroupI) {
}
func (w *SnakeWorld) RemoveObjGroup(id int64) {
}
func (w *SnakeWorld) NewObjGroup() ObjGroupI {
	og := ObjGroupBase{
		ID:       <-go4game.IdGenCh,
		GameObjs: make(map[int64]GameObjI),
		config:   w.config,
	}
	return &og
}

type ObjGroupBase struct {
	ID       int64
	GameObjs map[int64]GameObjI
	config   *SnakeConfig
}

func (og *ObjGroupBase) AddGameObj(GameObjI) {
}
func (og *ObjGroupBase) RemoveGameObj(id int64) {
}
func (og *ObjGroupBase) DoFrameAction(ftime time.Time) <-chan interface{} {
	rtn := make(chan interface{}, 1)
	return rtn
}

type GameObjBase struct {
	ID         int64
	GroupID    int64
	PosVector  go4game.Vector3D
	MoveVector go4game.Vector3D
}

func (o GameObjBase) String() string {
	return fmt.Sprintf("ID:%v Group%v", o.ID, o.GroupID)
}

func (o GameObjBase) Pos() go4game.Vector3D {
	return o.PosVector
}
