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
	chStep         <-chan []int
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
		t.makeMainObj()
	case *websocket.Conn:
		t.ClientConnInfo = *NewWsConnInfo(&t, conn.(*websocket.Conn))
	case AIActor:
		t.ClientConnInfo = *NewAIConnInfo(&t, conn.(AIActor))
		t.makeMainObj()
	default:
		log.Printf("unknown type %#v", conn)
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

func (t *Team) countObjByType(got GameObjectType) int {
	rtn := 0
	for _, v := range t.GameObjs {
		if v.ObjType == got {
			rtn++
		}
	}
	return rtn
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
	case <-time.After(GameConst.FrameRate):
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
			TeamInfo:  &TeamInfoPacket{SPObj: NewSPObj(t.findMainObj())},
		}
	case ReqFrameInfo:
		t.applyClientAction(ftime, p.ClientAct)
		rp = GamePacket{
			Cmd: RspFrameInfo,
			Spp: spp,
			TeamInfo: &TeamInfoPacket{
				SPObj:       NewSPObj(t.findMainObj()),
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

func (t *Team) actByTime(ftime time.Time, spp *SpatialPartition) []int {
	clist := make([]int, 0)
	for _, v := range t.GameObjs {
		clist = append(clist, v.ActByTime(ftime, spp)...)
	}
	for _, v := range t.GameObjs {
		if v.enabled == false {
			t.delGameObject(v)
			if v.ObjType == GameObjMain {
				t.makeMainObj()
				//t.addNewGameObject(v.ObjType, nil)
			}
		}
	}
	return clist
}

// 0(outer max) ~ GameConst.APIncFrame( 0,0,0)
func (t *Team) CalcAP(spp *SpatialPartition) int {
	o := t.findMainObj()
	if o == nil {
		return 0
	}
	l := o.PosVector.Abs()
	lm := spp.Size.Abs() / 2
	rtn := int((lm - l) / lm * float64(GameConst.APIncFrame))
	//log.Printf("ap:%v", rtn)
	return rtn
}

func (t *Team) doFrameWork(ftime time.Time, spp *SpatialPartition, w *WorldSerialize) <-chan []int {
	ap := t.CalcAP(spp)
	if ap < 0 {
		log.Printf("invalid ap team%v %v", t.ID, ap)
	}
	t.ActionPoint += ap

	chRtn := make(chan []int)
	go func() {
		rtn := t.processClientReq(ftime, w, spp)
		if !rtn {
			//chRtn <- false
			close(chRtn)
			return
		}
		chRtn <- t.actByTime(ftime, spp)
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

func (t *Team) makeMainObj() {
	if t.findMainObj() != nil {
		log.Printf("main obj exist %v", t)
		return
	}
	t.addNewGameObject(GameObjMain, nil)
	shieldcount := t.countObjByType(GameObjShield)
	for i := shieldcount; i < GameConst.ShieldCount; i++ {
		t.addNewGameObject(GameObjShield, nil)
	}
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
	case GameObjSuperBullet:
		mo := t.findMainObj()
		if mo != nil {
			o.MakeSuperBullet(mo, args.(*Vector3D))
		}
	case GameObjHommingBullet:
		mo := t.findMainObj()
		if mo != nil {
			targetid := args.([]int)[0]
			targetteamid := args.([]int)[1]
			o.MakeHommingBullet(mo, targetteamid, targetid)
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
	if act.HommingTargetID != nil {
		if t.ActionPoint >= GameConst.APHommingBullet {
			t.addNewGameObject(GameObjHommingBullet, act.HommingTargetID)
			t.ActionPoint -= GameConst.APHommingBullet
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use hommingbullet %v",
				t.ID, t.ActionPoint, act.HommingTargetID)
		}
	}
	if act.SuperBulletMv != nil {
		if t.ActionPoint >= GameConst.APSuperBullet {
			t.addNewGameObject(GameObjSuperBullet, act.SuperBulletMv)
			t.ActionPoint -= GameConst.APSuperBullet
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use superbullet %v",
				t.ID, t.ActionPoint, act.SuperBulletMv)
		}
	}
	return rtn
}
