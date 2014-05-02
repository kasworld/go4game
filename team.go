package go4game

import (
	//"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	//"reflect"
	"github.com/gorilla/websocket"
	"time"
)

type Team struct {
	ID             int
	PWorld         *World
	GameObjs       map[int]*GameObject
	ClientConnInfo ConnInfo
	spp            *SpatialPartition
	chStep         <-chan bool
	Color          int
}

func (m Team) String() string {
	return fmt.Sprintf("Team%v Objs:%v", m.ID, len(m.GameObjs))
}

func NewTeam(w *World, conn interface{}) *Team {
	t := Team{
		ID:       <-IdGenCh,
		PWorld:   w,
		GameObjs: make(map[int]*GameObject),
		Color:    rand.Intn(0x1000000),
	}
	switch conn.(type) {
	case net.Conn:
		t.ClientConnInfo = *NewTcpConnInfo(&t, conn.(net.Conn))
	case *websocket.Conn:
		t.ClientConnInfo = *NewWsConnInfo(&t, conn.(*websocket.Conn))
	case *AIConn:
		t.ClientConnInfo = *NewAIConnInfo(&t, conn.(*AIConn))
	default:
		log.Printf("unknown type %#v", conn)
	}
	t.addNewGameObject(GameObjMain, nil)

	for i := 0; i < 8; i++ {
		t.addNewGameObject(GameObjShield, nil)
	}
	for i := 0; i < 0; i++ {
		t.addNewGameObject(GameObjBullet, nil)
	}
	return &t
}

func (t *Team) findMainObj() *GameObject {
	for _, v := range t.GameObjs {
		if v.objType == GameObjMain {
			return v
		}
	}
	return nil
}

func (t *Team) processClientReq(ftime time.Time, w *WorldSerialize) bool {
	p, ok := <-t.ClientConnInfo.ReadCh
	if !ok { // read closed
		//log.Printf("client quit %v", t)
		return false
	}
	//log.Printf("client packet %v %v", t, p)
	var rp GamePacket
	switch p.Cmd {
	case ReqMakeTeam:
		rp = GamePacket{
			Cmd: RspMakeTeam,
		}
	case ReqWorldInfo:
		rp = GamePacket{
			Cmd:       RspWorldInfo,
			WorldInfo: w,
		}
	case ReqAIAct:
		t.applyClientAction(ftime, p.ClientAct)
		rp = GamePacket{
			Cmd: RspAIAct,
		}
	default:
		log.Printf("unknown packet %#v", p)
		return false
	}
	//log.Printf("client packet processed %v %v", t, rp.Cmd)
	t.ClientConnInfo.WriteCh <- &rp
	//log.Printf("end processClientReq %v", t)
	return true
}

// work fn
func (t *Team) actByTime(ftime time.Time, spp *SpatialPartition) bool {
	for _, v := range t.GameObjs {
		v.ActByTime(ftime, spp)
	}
	for _, v := range t.GameObjs {
		if v.enabled == false {
			t.delGameObject(v)
			if v.objType == GameObjMain {
				t.addNewGameObject(v.objType, nil)
			}
			if v.objType == GameObjShield {
				t.addNewGameObject(v.objType, nil)
			}

		}
	}
	return true
}

func (t *Team) doFrameWork(ftime time.Time, spp *SpatialPartition, w *WorldSerialize) <-chan bool {
	chRtn := make(chan bool)
	go func() {
		rtn := t.processClientReq(ftime, w)
		if !rtn {
			chRtn <- false
			return
		}
		t.actByTime(ftime, spp)
		chRtn <- true
	}()
	return chRtn
}

func (t *Team) endTeam() {
	close(t.ClientConnInfo.WriteCh) // stop writeloop
	if t.ClientConnInfo.Conn != nil {
		t.ClientConnInfo.Conn.Close() // stop read loop
	}
	if t.ClientConnInfo.WsConn != nil {
		t.ClientConnInfo.WsConn.Close() // stop read loop
	}
	//log.Printf("team end %v", t)
}

func (t *Team) addNewGameObject(objType GameObjectType, args interface{}) *GameObject {
	o := NewGameObject(t)
	switch objType {
	case GameObjMain:
		o.MakeMainObj()
	case GameObjShield:
		mo := t.findMainObj()
		if mo != nil {
			o.MakeShield(mo)
		}
	case GameObjBullet:
		mo := t.findMainObj()
		if mo != nil {
			o.MakeBullet(mo, args.(*Vector3D))
		}
	default:
		log.Printf("invalid GameObjectType %v", t)
		return nil
	}
	t.GameObjs[o.ID] = o
	return o
}

func (t *Team) delGameObject(o *GameObject) {
	delete(t.GameObjs, o.ID)
}

func (t *Team) applyClientAction(ftime time.Time, act *ClientActionPacket) {
	if act == nil {
		return
	}
	mo := t.findMainObj()
	if act.Accel != nil {
		mo.accelVector.Add(act.Accel)
	}
	if act.NormalBulletMv != nil {
		t.addNewGameObject(GameObjBullet, act.NormalBulletMv)
	}
	if act.HommingTargetID != 0 {
	}
	if act.SuperBulletMv != nil {
	}

}
