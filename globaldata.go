package go4game

import (
	"time"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize = 0xffff              // Maximum message size allowed from peer.
)

// client conn type
type ClientType int

const (
	_ ClientType = iota
	TCPClient
	WebSockClient
	AIClient
)

// packet type
type PacketType int

const (
	_ PacketType = iota
	ReqFrameInfo
	RspFrameInfo
	ReqWorldInfo
	RspWorldInfo
	ReqAIAct
	RspAIAct
)

type GameObjectType int

const (
	_ GameObjectType = iota
	GameObjMain
	GameObjShield
	GameObjBullet
)

type gameConst struct {
	WorldMax        Vector3D
	WorldMin        Vector3D
	MaxObjectRadius float64
}

var GameConst gameConst

func init() {
	GameConst = gameConst{
		WorldMin:        Vector3D{-500, -500, -500},
		WorldMax:        Vector3D{500, 500, 500},
		MaxObjectRadius: 10,
	}
}
