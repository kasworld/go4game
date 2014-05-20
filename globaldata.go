package go4game

import (
	"log"
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
	MoveLimit [GameObjEnd]float64
	Radius    [GameObjEnd]float64
}{
	MoveLimit: [GameObjEnd]float64{
		GameObjMain: 100, GameObjShield: 200, GameObjBullet: 300, GameObjHommingBullet: 200, GameObjSuperBullet: 600},
	Radius: [GameObjEnd]float64{
		GameObjMain: 10, GameObjShield: 5, GameObjBullet: 5, GameObjHommingBullet: 7, GameObjSuperBullet: 15},
}

// harmed obj : can harm obj
var InteractionMap = [GameObjEnd][GameObjEnd]bool{
	GameObjMain: [GameObjEnd]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: true, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjShield: [GameObjEnd]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: true, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjBullet: [GameObjEnd]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: true, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjHommingBullet: [GameObjEnd]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: false, GameObjHommingBullet: true, GameObjSuperBullet: true},
	GameObjSuperBullet: [GameObjEnd]bool{
		GameObjMain: true, GameObjShield: true, GameObjBullet: false, GameObjHommingBullet: true, GameObjSuperBullet: true},
}

const WorldSize = 500

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
	APIncFrame           int
	KillScore            int
	ShieldCount          int
	MaxObjectRadius      float64
}{
	TcpListen:            "0.0.0.0:6666",
	WsListen:             "0.0.0.0:8080",
	WorldMin:             Vector3D{-WorldSize, -WorldSize, -WorldSize},
	WorldMax:             Vector3D{WorldSize, WorldSize, WorldSize},
	FramePerSec:          60.0,
	RemoveEmptyWorld:     false,
	MaxTcpClientPerWorld: 32,
	MaxWsClientPerWorld:  32,
	StartWorldCount:      1,
	NpcCountPerWorld:     32,
	ClearY:               false,
	APIncFrame:           10,
	KillScore:            1,
	ShieldCount:          8,
	MaxObjectRadius:      1, // changed by init
}

type ClientActionType int

const (
	ActionAccel ClientActionType = iota
	ActionBullet
	ActionSuperBullet
	ActionHommingBullet
	ActionBurstBullet
	ActionEnd
)

var ActionPoints = [ActionEnd]int{
	ActionAccel:         1,
	ActionBullet:        10,
	ActionSuperBullet:   80,
	ActionHommingBullet: 100,
	ActionBurstBullet:   10,
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
	log.Printf("%#v", ObjDefault)
}
