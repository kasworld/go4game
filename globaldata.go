package go4game

import (
	"log"
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

type GameObjectType int

const (
	_ GameObjectType = iota
	GameObjMain
	GameObjShield
	GameObjBullet
)

var GameConst = struct {
	WorldMax        Vector3D
	WorldMin        Vector3D
	MaxObjectRadius float64
	APAccel         int
	APBullet        int
	APBurstShot     int
	APHommingBullet int
	APSuperBullet   int
	APIncFrame      int
	KillScore       int
}{
	WorldMin:        Vector3D{-500, -500, -500},
	WorldMax:        Vector3D{500, 500, 500},
	MaxObjectRadius: 10,

	APAccel:         5,
	APBullet:        20,
	APBurstShot:     10,
	APHommingBullet: 40,
	APSuperBullet:   40,

	APIncFrame: 10,
	KillScore:  1,
}

var InteractionMap = map[GameObjectType]map[GameObjectType]bool{
	GameObjMain:   map[GameObjectType]bool{GameObjMain: true, GameObjShield: true, GameObjBullet: true},
	GameObjShield: map[GameObjectType]bool{GameObjMain: true, GameObjShield: true, GameObjBullet: true},
	GameObjBullet: map[GameObjectType]bool{GameObjMain: true, GameObjShield: true, GameObjBullet: true},
}

func init() {
	log.Printf("%#v", GameConst)
	log.Printf("%#v", InteractionMap)
}
