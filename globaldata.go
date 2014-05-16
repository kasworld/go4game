package go4game

import (
//"log"
//"time"
)

type GameObjectType int

const (
	_ GameObjectType = iota
	GameObjMain
	GameObjShield
	GameObjBullet
	GameObjHommingBullet
	GameObjSuperBullet
	GameObjEnd
)

var ObjDefault = struct {
	MoveLimit map[GameObjectType]float64
	Radius    map[GameObjectType]float64
}{
	MoveLimit: map[GameObjectType]float64{
		GameObjMain: 100, GameObjShield: 200, GameObjBullet: 300, GameObjHommingBullet: 200, GameObjSuperBullet: 600},
	Radius: map[GameObjectType]float64{
		GameObjMain: 10, GameObjShield: 5, GameObjBullet: 5, GameObjHommingBullet: 10, GameObjSuperBullet: 20},
}

// harmed obj : can harm obj
var InteractionMap = map[GameObjectType]map[GameObjectType]bool{
	GameObjMain: map[GameObjectType]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: true, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjShield: map[GameObjectType]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: true, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjBullet: map[GameObjectType]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: true, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjHommingBullet: map[GameObjectType]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: false, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjSuperBullet: map[GameObjectType]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: false, GameObjHommingBullet: true, GameObjSuperBullet: true},
}

var GameConst = struct {
	TcpListen            string
	WsListen             string
	ClearY               bool
	FramePerSec          float64
	NpcCountPerWorld     int
	MaxTcpClientPerWorld int
	MaxWsClientPerWorld  int
	StartWorldCount      int
	RemoveEmptyWorld     bool
	WorldMax             Vector3D
	WorldMin             Vector3D

	APAccel         int
	APBullet        int
	APBurstShot     int
	APHommingBullet int
	APSuperBullet   int
	APIncFrame      int

	KillScore       int
	ShieldCount     int
	MaxObjectRadius float64
}{
	TcpListen: "0.0.0.0:6666",
	WsListen:  "0.0.0.0:8080",
	WorldMin:  Vector3D{-500, -500, -500},
	WorldMax:  Vector3D{500, 500, 500},

	FramePerSec:          60.0,
	RemoveEmptyWorld:     false,
	MaxTcpClientPerWorld: 32,
	MaxWsClientPerWorld:  32,
	StartWorldCount:      1,
	NpcCountPerWorld:     1024,
	ClearY:               false,

	APAccel:         1,
	APBullet:        10,
	APBurstShot:     10,
	APHommingBullet: 100,
	APSuperBullet:   80,
	APIncFrame:      10,

	KillScore:       1,
	ShieldCount:     8,
	MaxObjectRadius: 20, // changed by init
}

var ObjSqd [GameObjEnd][GameObjEnd]float64

func init() {
	for _, o := range ObjDefault.Radius {
		if o > GameConst.MaxObjectRadius {
			GameConst.MaxObjectRadius = o
		}
	}
	for o1 := GameObjMain; o1 < GameObjEnd; o1++ {
		for o2 := GameObjMain; o2 < GameObjEnd; o2++ {
			ObjSqd[o1][o2] = (ObjDefault.Radius[o1] + ObjDefault.Radius[o2]) * (ObjDefault.Radius[o1] + ObjDefault.Radius[o2])
		}
	}
	//log.Printf("%#v", ObjSqd)
}
