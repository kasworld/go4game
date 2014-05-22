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
	HomePos     Vector3D
	MainObjID   int64

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

type TeamInfo struct {
	ID         int64
	ClientInfo string
	Objs       int
	AP         int
	PacketStat string
	CollStat   string
	Color      int
	FontColor  int
	Score      int
}

func (t *Team) NewTeamInfo() *TeamInfo {
	return &TeamInfo{
		ID:         t.ID,
		ClientInfo: t.ClientConnInfo.String(),
		Objs:       len(t.GameObjs),
		AP:         t.ActionPoint,
		PacketStat: t.PacketStat.String(),
		CollStat:   t.CollisionStat.String(),
		Color:      t.Color,
		FontColor:  0xffffff ^ t.Color,
		Score:      t.Score,
	}
}

type ByScore []TeamInfo

func (s ByScore) Len() int {
	return len(s)
}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByScore) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

func NewTeam(w *World, conn interface{}) *Team {
	t := Team{
		ID:            <-IdGenCh,
		GameObjs:      make(map[int64]*GameObject),
		Color:         rand.Intn(0x1000000),
		PacketStat:    *NewActionStat(),
		CollisionStat: *NewActionStat(),
	}
	switch conn.(type) {
	case net.Conn:
		t.ClientConnInfo = *NewTcpConnInfo(conn.(net.Conn))
		t.makeMainObj()
	case *websocket.Conn:
		t.ClientConnInfo = *NewWsConnInfo(conn.(*websocket.Conn))
	case AIActor:
		t.ClientConnInfo = *NewAIConnInfo(conn.(AIActor))
		t.makeMainObj()
	default:
		log.Printf("unknown type %#v", conn)
	}
	t.HomePos = GameConst.WorldCube.RandVector().Idiv(2)
	if GameConst.ClearY {
		t.HomePos[1] = 0
	}
	return &t
}

func (t *Team) moveHomePos() {
	t.HomePos = t.HomePos.Add(GameConst.WorldCube.RandVector().Idiv(100))
	for i, v := range t.HomePos {
		if v > GameConst.WorldCube.Max[i] {
			t.HomePos[i] = GameConst.WorldCube.Max[i]
		}
		if v < GameConst.WorldCube.Min[i] {
			t.HomePos[i] = GameConst.WorldCube.Min[i]
		}
	}
	if GameConst.ClearY {
		t.HomePos[1] = 0
	}
}

func (t *Team) findMainObj() *GameObject {
	return t.GameObjs[t.MainObjID]
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
				HomePos:     t.HomePos,
			},
		}
	default:
		log.Printf("unknown packet %#v", p)
		return false
	}
	t.ClientConnInfo.WriteCh <- &rp
	return true
}

func (t *Team) doFrameWork(world *World, ftime time.Time) <-chan IDList {
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
			delete(t.GameObjs, id)
			if v.ObjType == GameObjMain {
				t.makeMainObj()
				t.Score -= GameConst.KillScore
			}
		}
	}
	t.moveHomePos()
	return clist
}

// 0(outer max) ~ GameConst.APIncFrame( 0,0,0)
func (t *Team) CalcAP(spp *SpatialPartition) int {
	o := t.findMainObj()
	if o == nil {
		return 0
	}
	lenToHomepos := o.PosVector.LenTo(t.HomePos)
	lmax := spp.Size.Abs()
	rtn := int((lmax - lenToHomepos) / lmax * float64(GameConst.APIncFrame))
	//log.Printf("ap:%v", rtn)
	return rtn
}

func (t *Team) endTeam() {
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
	t.addNewGameObject(GameObjMain, nil)
	shieldcount := t.countObjByType(GameObjShield)
	for i := shieldcount; i < GameConst.ShieldCount; i++ {
		t.addNewGameObject(GameObjShield, nil)
	}
}

func (t *Team) addNewGameObject(ObjType GameObjectType, args interface{}) *GameObject {
	o := NewGameObject(t.ID)
	switch ObjType {
	case GameObjMain:
		o.MakeMainObj()
		t.MainObjID = o.ID
	case GameObjShield:
		mo := t.findMainObj()
		if mo != nil {
			o.MakeShield(mo)
		}
	case GameObjBullet:
		mo := t.findMainObj()
		if mo != nil {
			o.MakeBullet(mo, args.(Vector3D))
		}
	case GameObjSuperBullet:
		mo := t.findMainObj()
		if mo != nil {
			o.MakeSuperBullet(mo, args.(Vector3D))
		}
	case GameObjHommingBullet:
		mo := t.findMainObj()
		if mo != nil {
			targetid := args.(IDList)[0]
			targetteamid := args.(IDList)[1]
			o.MakeHommingBullet(mo, targetteamid, targetid)
		}
	default:
		log.Printf("invalid GameObjectType %v", t)
		return nil
	}
	t.GameObjs[o.ID] = o
	return o
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
			t.addNewGameObject(GameObjBullet, *act.NormalBulletMv)
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
				t.addNewGameObject(GameObjBullet, RandVector3D(-300, 300))
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
			t.addNewGameObject(GameObjHommingBullet, act.HommingTargetID)
			t.ActionPoint -= GameConst.AP[ActionHommingBullet]
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use hommingbullet %v",
				t.ID, t.ActionPoint, act.HommingTargetID)
		}
	}
	if act.SuperBulletMv != nil {
		if t.ActionPoint >= GameConst.AP[ActionSuperBullet] {
			t.addNewGameObject(GameObjSuperBullet, *act.SuperBulletMv)
			t.ActionPoint -= GameConst.AP[ActionSuperBullet]
			rtn++
		} else {
			log.Printf("Team%v ap:%v over use superbullet %v",
				t.ID, t.ActionPoint, act.SuperBulletMv)
		}
	}
	return rtn
}
