package go4game

import (
	"time"
)

func (o GameObject) String() string {
	return o.SPObj.String()
}

type GameObject struct {
	*SPObj

	colcount     int64
	enabled      bool
	startTime    time.Time
	endTime      time.Time
	accelVector  Vector3D
	targetObjID  int64
	targetTeamID int64
	lastMoveTime time.Time

	moveByTimeFn      GameObjectActFn
	borderActionFn    GameObjectActFn
	collisionActionFn GameObjectActFn
	expireActionFn    GameObjectActFn
}

func NewGameObject(teamID int64) *GameObject {
	o := GameObject{
		SPObj: &SPObj{
			ID:     <-IdGenCh,
			TeamID: teamID,
		},
		enabled:           true,
		startTime:         time.Now(),
		lastMoveTime:      time.Now(),
		moveByTimeFn:      moveByTimeFn_accel,
		borderActionFn:    borderActionFn_Bounce,
		collisionActionFn: collisionFn_default,
		expireActionFn:    expireFn_default,
	}
	return &o
}

func (o *GameObject) ToSPObj() *SPObj {
	if o == nil {
		return nil
	}
	rtn := *o.SPObj
	return &rtn
}

type ActionFnEnvInfo struct {
	frameTime time.Time
	world     *World
	o         *GameObject
	clist     IDList
}

func (e *ActionFnEnvInfo) doPartMainObj(v *SPObj) bool {
	e.o.colcount++
	if (e.o.TeamID != v.TeamID) && e.o.SPObj.IsCollision(v) {
		e.clist = append(e.clist, v.TeamID)
	}
	return false
}

func (e *ActionFnEnvInfo) doPartOtherObj(v *SPObj) bool {
	e.o.colcount++
	if (e.o.TeamID != v.TeamID) && e.o.SPObj.IsCollision(v) {
		return true
	}
	return false
}

func (o *GameObject) ActByTime(world *World, t time.Time) IDList {
	defer func() {
		o.lastMoveTime = t
	}()

	envInfo := ActionFnEnvInfo{
		frameTime: t,
		world:     world,
		o:         o,
		clist:     make(IDList, 0),
	}
	// check expire
	if !o.endTime.IsZero() && o.endTime.Before(t) {
		if o.expireActionFn != nil {
			ok := o.expireActionFn(o, &envInfo)
			if ok != true {
				return envInfo.clist
			}
		}
	}

	var isCollision bool
	hr := NewHyperRectByCR(o.PosVector, GameConst.Radius[o.ObjType]+GameConst.MaxObjectRadius)
	if o.ObjType == GameObjMain {
		envInfo.world.octree.QueryByHyperRect(envInfo.doPartMainObj, hr)
	} else {
		isCollision = envInfo.world.octree.QueryByHyperRect(envInfo.doPartOtherObj, hr)
	}

	if isCollision || len(envInfo.clist) > 0 {
		if o.collisionActionFn != nil {
			ok := o.collisionActionFn(o, &envInfo)
			if ok != true {
				return envInfo.clist
			}
		}
	}
	if o.moveByTimeFn != nil {
		// change PosVector by movevector
		ok := o.moveByTimeFn(o, &envInfo)
		if ok != true {
			return envInfo.clist
		}
	}
	if o.borderActionFn != nil {
		// check wall action ( wrap, bounce )
		ok := o.borderActionFn(o, &envInfo)
		if ok != true {
			return envInfo.clist
		}
	}
	return envInfo.clist
}

type GameObjectActFn func(m *GameObject, envInfo *ActionFnEnvInfo) bool

func expireFn_default(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	m.enabled = false
	return false
}

func collisionFn_default(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	m.enabled = false
	return false
}

func moveByTimeFn_none(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	return true
}

func moveByTimeFn_accel(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	dur := float64(envInfo.frameTime.Sub(m.lastMoveTime)) / float64(time.Second)
	m.MoveVector = m.MoveVector.Add(m.accelVector.Imul(dur))
	if m.MoveVector.Abs() > GameConst.MoveLimit[m.ObjType] {
		m.MoveVector = m.MoveVector.NormalizedTo(GameConst.MoveLimit[m.ObjType])
	}
	m.PosVector = m.PosVector.Add(m.MoveVector.Imul(dur))
	return true
}

func moveByTimeFn_home(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	dur := float64(envInfo.frameTime.Sub(m.lastMoveTime)) / float64(time.Second)
	m.accelVector = GameConst.WorldCube.RandVector()
	m.MoveVector = m.MoveVector.Add(m.accelVector.Imul(dur))
	if m.MoveVector.Abs() > GameConst.MoveLimit[m.ObjType] {
		m.MoveVector = m.MoveVector.NormalizedTo(GameConst.MoveLimit[m.ObjType])
	}
	m.PosVector = m.PosVector.Add(m.MoveVector.Imul(dur))
	return true
}

func moveByTimeFn_shield(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	mo := envInfo.world.Teams[m.TeamID].findMainObj()
	if mo == nil {
		return false
	}
	dur := float64(envInfo.frameTime.Sub(m.startTime)) / float64(time.Second)
	//axis := &Vector3D{0, math.Copysign(20, m.MoveVector[0]), 0}
	axis := mo.MoveVector // .NormalizedTo(20)
	//p := m.accelVector.NormalizedTo(20)
	p := mo.MoveVector.Cross(m.MoveVector).NormalizedTo(20)
	m.PosVector = mo.PosVector.Add(p.RotateAround(axis, dur+m.accelVector.Abs()))
	return true
}

func moveByTimeFn_clock(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	dur := float64(envInfo.frameTime.Sub(m.startTime)) / float64(time.Second)
	p := m.MoveVector.Cross(m.accelVector)
	//m.PosVector = m.MoveVector.NormalizedTo(30).RotateAround(p, dur*5)
	m.PosVector = m.MoveVector.RotateAround(p, dur)
	return true
}

func moveByTimeFn_homming(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	// how to other team obj pos? without panic
	targetTeam := envInfo.world.Teams[m.targetTeamID]
	if targetTeam == nil {
		m.enabled = false
		return false
	}
	targetobj := targetTeam.GameObjs[m.targetObjID]
	if targetobj == nil {
		m.enabled = false
		return false
	}
	m.accelVector = targetobj.PosVector.Sub(m.PosVector).NormalizedTo(GameConst.MoveLimit[m.ObjType])
	return moveByTimeFn_accel(m, envInfo)
}

func borderActionFn_Bounce(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > GameConst.WorldCube.Max[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.MoveVector[i] = -m.MoveVector[i]
			m.PosVector[i] = GameConst.WorldCube.Max[i]
		}
		if m.PosVector[i] < GameConst.WorldCube.Min[i] {
			m.accelVector[i] = -m.accelVector[i]
			m.MoveVector[i] = -m.MoveVector[i]
			m.PosVector[i] = GameConst.WorldCube.Min[i]
		}
	}
	return true
}

func borderActionFn_Block(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > GameConst.WorldCube.Max[i] {
			m.accelVector[i] = 0
			m.MoveVector[i] = 0
			m.PosVector[i] = GameConst.WorldCube.Max[i]
		}
		if m.PosVector[i] < GameConst.WorldCube.Min[i] {
			m.accelVector[i] = 0
			m.MoveVector[i] = 0
			m.PosVector[i] = GameConst.WorldCube.Min[i]
		}
	}
	return true
}

func borderActionFn_Wrap(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if m.PosVector[i] > GameConst.WorldCube.Max[i] {
			m.PosVector[i] = GameConst.WorldCube.Min[i]
		}
		if m.PosVector[i] < GameConst.WorldCube.Min[i] {
			m.PosVector[i] = GameConst.WorldCube.Max[i]
		}
	}
	return true
}

func borderActionFn_Disable(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if (m.PosVector[i] > GameConst.WorldCube.Max[i]) || (m.PosVector[i] < GameConst.WorldCube.Min[i]) {
			m.enabled = false
			return false
		}
	}
	return true
}

func borderActionFn_B2_Disable(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	for i := range m.PosVector {
		if (m.PosVector[i] > GameConst.WorldCube2.Max[i]) || (m.PosVector[i] < GameConst.WorldCube2.Min[i]) {
			m.enabled = false
			return false
		}
	}
	return true
}

func borderActionFn_None(m *GameObject, envInfo *ActionFnEnvInfo) bool {
	return true
}
