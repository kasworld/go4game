package go4game

import (
	//"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net"
	"time"
)

type Team struct {
	ID          int64
	Color       int
	ActionPoint int
	Score       int
	MainObjID   int64
	HomeObjID   int64

	PacketStat     ActionStat
	CollisionStat  ActionStat
	GameObjs       map[int64]*GameObject
	ClientConnInfo ConnInfo
	chStep         <-chan IDList
}

func (m Team) String() string {
	return fmt.Sprintf("Team%v %v Objs:%v Score:%v AP:%v, PacketStat:%v, Coll:%v",
		m.ID, m.ClientConnInfo, len(m.GameObjs), m.Score, m.ActionPoint, m.PacketStat, m.CollisionStat)
}

func NewTeam(conn interface{}) *Team {
	t := Team{
		ID:            <-IdGenCh,
		GameObjs:      make(map[int64]*GameObject, 10),
		Color:         rand.Intn(0x1000000),
		PacketStat:    *NewActionStat(),
		CollisionStat: *NewActionStat(),
	}
	switch conn.(type) {
	default:
		log.Printf("unknown type %#v", conn)
	case net.Conn:
		t.ClientConnInfo = *NewTcpConnInfo(conn.(net.Conn))
		t.makeMainObj()
	case *websocket.Conn:
		t.ClientConnInfo = *NewWsConnInfo(conn.(*websocket.Conn))
	case AIActor:
		t.ClientConnInfo = *NewAIConnInfo(conn.(AIActor))
		t.makeMainObj()
	}
	o := t.addObject(NewGameObject(t.ID).MakeHomeMarkObj())
	t.HomeObjID = o.ID

	t.addDeco()
	return &t
}

func (t *Team) addDeco() {
	avt := GameConst.WorldCube.RandVector().Idiv(10)
	mvvt := GameConst.WorldCube.RandVector().Idiv(10)
	for i := 0; i < 50; i++ {
		o := NewGameObject(t.ID).MakeDecoObj()

		o.accelVector = avt //.NormalizedTo(float64(i * 10+1))
		o.MoveVector = mvvt.NormalizedTo(float64(i*16 + 1))
		t.addObject(o)
	}
}

func (t *Team) addObject(o *GameObject) *GameObject {
	t.GameObjs[o.ID] = o
	return o
}

func (t *Team) removeObject(id int64) {
	delete(t.GameObjs, id)
}

func (t *Team) findMainObj() *GameObject {
	return t.GameObjs[t.MainObjID]
}
func (t *Team) findHomeObj() *GameObject {
	return t.GameObjs[t.HomeObjID]
}

func (t *Team) countObjByType(got GameObjectType) int {
	rtn := 0
	for _, v := range t.GameObjs {
		if v != nil && v.ObjType == got {
			rtn++
		}
	}
	return rtn
}

func (t *Team) processClientReq(ftime time.Time, w *WorldDisp, spp *SpatialPartition) bool {
	var p *GamePacket
	var ok bool
	select {
	case p, ok = <-t.ClientConnInfo.ReadCh:
		if !ok { // read closed
			return false
		}
	case <-time.After(time.Duration(1000/GameConst.FramePerSec) * time.Millisecond):
	}
	if p == nil {
		return true
	}
	t.PacketStat.Inc()
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
				HomePos:     t.findHomeObj().PosVector,
			},
		}
	default:
		log.Printf("unknown packet %#v", p)
		return false
	}
	t.ClientConnInfo.WriteCh <- &rp
	return true
}

func (t *Team) Do1Frame(world *World, ftime time.Time) <-chan IDList {
	spp := world.spp
	w := world.worldSerial
	ap := t.CalcAP(spp)
	if ap < 0 {
		log.Printf("invalid ap team%v %v", t.ID, ap)
	}
	t.ActionPoint += ap

	chRtn := make(chan IDList)
	go func() {
		rtn := t.processClientReq(ftime, w, spp)
		if !rtn {
			close(chRtn)
			return
		}
		chRtn <- t.actByTime(world, ftime)
	}()
	return chRtn
}

func (t *Team) actByTime(world *World, ftime time.Time) IDList {
	clist := make(IDList, 0)
	for _, v := range t.GameObjs {
		v.colcount = 0
		clist = append(clist, v.ActByTime(world, ftime)...)
		t.CollisionStat.Add(v.colcount)
	}
	for id, v := range t.GameObjs {
		if v.enabled == false {
			t.removeObject(id)
			if v.ObjType == GameObjMain {
				//t.makeMainObj()
				t.Score -= GameConst.KillScore
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
	homepos := t.findHomeObj().PosVector
	lenToHomepos := o.PosVector.LenTo(homepos)
	lmax := spp.Size.Abs()
	rtn := int((lmax - lenToHomepos) / lmax * float64(GameConst.APIncFrame))
	//log.Printf("ap:%v", rtn)
	return rtn
}

func (t *Team) endTeam() {
	if t == nil {
		log.Printf("warning end nil team")
		return
	}
	//t.Status = false
	//log.Printf("end team %v", t.ID)
	close(t.ClientConnInfo.WriteCh) // stop writeloop
	if t.ClientConnInfo.Conn != nil {
		t.ClientConnInfo.Conn.Close() // stop read loop
	}
	if t.ClientConnInfo.WsConn != nil {
		t.ClientConnInfo.WsConn.Close() // stop read loop
	}
}

func (t *Team) makeMainObj() {
	if t.findMainObj() != nil {
		log.Printf("main obj exist %v", t)
		return
	}
	mo := t.addObject(NewGameObject(t.ID).MakeMainObj())
	t.MainObjID = mo.ID

	shieldcount := t.countObjByType(GameObjShield)
	for i := shieldcount; i < GameConst.ShieldCount; i++ {
		t.addObject(NewGameObject(t.ID).MakeShield(mo))
	}
}

func (t *Team) fireBullet(ObjType GameObjectType, args interface{}) *GameObject {
	mo := t.findMainObj()
	if mo == nil {
		return nil
	}
	o := NewGameObject(t.ID)
	switch ObjType {
	default:
		log.Printf("invalid GameObjectType %v", t)
		return nil
	case GameObjBullet:
		o.MakeBullet(mo, args.(Vector3D))
	case GameObjSuperBullet:
		o.MakeSuperBullet(mo, args.(Vector3D))
	case GameObjHommingBullet:
		targetid := args.(IDList)[0]
		targetteamid := args.(IDList)[1]
		o.MakeHommingBullet(mo, targetteamid, targetid)
	}
	return t.addObject(o)
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
		if t.ActionPoint >= GameConst.AP[ActionAccel] {
			mo.accelVector = *act.Accel
			t.ActionPoint -= GameConst.AP[ActionAccel]
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use accel %v",
				t.ID, t.ActionPoint, act.Accel)
		}

	}
	if act.NormalBulletMv != nil {
		if t.ActionPoint >= GameConst.AP[ActionBullet] {
			t.fireBullet(GameObjBullet, *act.NormalBulletMv)
			t.ActionPoint -= GameConst.AP[ActionBullet]
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use bullet %v",
				t.ID, t.ActionPoint, act.NormalBulletMv)
		}
	}
	if act.BurstShot > 0 {
		if t.ActionPoint >= act.BurstShot*GameConst.AP[ActionBurstBullet] {
			for i := 0; i < act.BurstShot; i++ {
				t.fireBullet(GameObjBullet, RandVector3D(-300, 300))
			}
			t.ActionPoint -= GameConst.AP[ActionBurstBullet] * act.BurstShot
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use burstbullet %v",
				t.ID, t.ActionPoint, act.BurstShot)
		}
	}
	if act.HommingTargetID != nil {
		if t.ActionPoint >= GameConst.AP[ActionHommingBullet] {
			t.fireBullet(GameObjHommingBullet, act.HommingTargetID)
			t.ActionPoint -= GameConst.AP[ActionHommingBullet]
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use hommingbullet %v",
				t.ID, t.ActionPoint, act.HommingTargetID)
		}
	}
	if act.SuperBulletMv != nil {
		if t.ActionPoint >= GameConst.AP[ActionSuperBullet] {
			t.fireBullet(GameObjSuperBullet, *act.SuperBulletMv)
			t.ActionPoint -= GameConst.AP[ActionSuperBullet]
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use superbullet %v",
				t.ID, t.ActionPoint, act.SuperBulletMv)
		}
	}
	return rtn
}
