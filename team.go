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
	chStep         <-chan bool
	Color          int
	PacketStat     ActionStat
	ActionPoint    int
	Score          int
}

func (m Team) String() string {
	return fmt.Sprintf("Team%v Objs:%v Score:%v AP:%v, PacketStat:%v",
		m.ID, len(m.GameObjs), m.Score, m.ActionPoint, m.PacketStat)
}

func NewTeam(w *World, conn interface{}) *Team {
	t := Team{
		ID:         <-IdGenCh,
		PWorld:     w,
		GameObjs:   make(map[int]*GameObject),
		Color:      rand.Intn(0x1000000),
		PacketStat: *NewActionStat(),
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
		if v.ObjType == GameObjMain {
			return v
		}
	}
	return nil
}

func (t *Team) processClientReq(ftime time.Time, w *WorldSerialize, spp *SpatialPartition) bool {
	var p *GamePacket
	var ok bool
	select {
	case p, ok = <-t.ClientConnInfo.ReadCh:
		if !ok { // read closed
			//log.Printf("client quit %v", t)
			return false
		}
	case <-time.After(1000 / 60 * time.Millisecond):
	}
	if p == nil {
		log.Printf("timeout team%v", t.ID)
		return true
	}
	t.PacketStat.Inc()
	//log.Printf("client packet %v %v", t, p)
	var rp GamePacket
	switch p.Cmd {
	case ReqWorldInfo:
		rp = GamePacket{
			Cmd:       RspWorldInfo,
			WorldInfo: w,
			TeamInfo:  &TeamInfoPacket{SPObj: *NewSPObj(t.findMainObj())},
		}
	case ReqFrameInfo:
		t.applyClientAction(ftime, p.ClientAct)
		rp = GamePacket{
			Cmd: RspFrameInfo,
			Spp: spp,
			TeamInfo: &TeamInfoPacket{
				SPObj:       *NewSPObj(t.findMainObj()),
				ActionPoint: t.ActionPoint,
				Score:       t.Score,
			},
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

func (t *Team) actByTime(ftime time.Time, spp *SpatialPartition) bool {
	for _, v := range t.GameObjs {
		v.ActByTime(ftime, spp)
	}
	for _, v := range t.GameObjs {
		if v.enabled == false {
			t.delGameObject(v)
			if v.ObjType == GameObjMain {
				t.addNewGameObject(v.ObjType, nil)
				t.Score -= GameConst.KillScore
			}
			if v.ObjType == GameObjShield {
				t.addNewGameObject(v.ObjType, nil)
			}

		}
	}
	return true
}

// 0(outer max) ~ GameConst.APIncFrame( 0,0,0)
func (t *Team) CalcAP(spp *SpatialPartition) int {
	o := t.findMainObj()
	l := o.PosVector.Abs()
	lm := spp.Size.Abs() / 2
	rtn := int((lm - l) / lm * float64(GameConst.APIncFrame))
	//log.Printf("ap:%v", rtn)
	return rtn
}

func (t *Team) doFrameWork(ftime time.Time, spp *SpatialPartition, w *WorldSerialize) <-chan bool {
	ap := t.CalcAP(spp)
	if ap < 0 {
		log.Printf("invalid ap team%v %v", t.ID, ap)
	}
	t.ActionPoint += ap
	t.Score += 1

	chRtn := make(chan bool)
	go func() {
		rtn := t.processClientReq(ftime, w, spp)
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

func (t *Team) addNewGameObject(ObjType GameObjectType, args interface{}) *GameObject {
	o := NewGameObject(t)
	switch ObjType {
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

func (t *Team) applyClientAction(ftime time.Time, act *ClientActionPacket) int {
	rtn := 0
	if act == nil {
		return rtn
	}
	mo := t.findMainObj()
	if mo == nil {
		return rtn
	}
	if act.Accel != nil {
		if t.ActionPoint >= GameConst.APAccel {
			mo.accelVector = *act.Accel
			t.ActionPoint -= GameConst.APAccel
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use accel %v",
				t.ID, t.ActionPoint, act.Accel)
		}

	}
	if act.NormalBulletMv != nil {
		if t.ActionPoint >= GameConst.APBullet {
			t.addNewGameObject(GameObjBullet, act.NormalBulletMv)
			t.ActionPoint -= GameConst.APBullet
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use bullet %v",
				t.ID, t.ActionPoint, act.NormalBulletMv)
		}
	}
	if act.BurstShot > 0 {
		if t.ActionPoint >= act.BurstShot*GameConst.APBurstShot {
			for i := 0; i < act.BurstShot; i++ {
				t.addNewGameObject(GameObjBullet, RandVector3D(-300, 300))
			}
			t.ActionPoint -= GameConst.APBurstShot * act.BurstShot
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use burstbullet %v",
				t.ID, t.ActionPoint, act.BurstShot)
		}
	}
	if act.HommingTargetID != 0 {
		if t.ActionPoint >= GameConst.APHommingBullet {
			t.ActionPoint -= GameConst.APHommingBullet
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use hommingbullet %v",
				t.ID, t.ActionPoint, act.HommingTargetID)
		}
	}
	if act.SuperBulletMv != nil {
		if t.ActionPoint >= GameConst.APSuperBullet {
			t.ActionPoint -= GameConst.APSuperBullet
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use superbullet %v",
				t.ID, t.ActionPoint, act.SuperBulletMv)
		}
	}
	return rtn
}
