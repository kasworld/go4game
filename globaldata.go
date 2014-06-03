package go4game

import (
	"log"
	//"time"
	"encoding/json"
	"math"
	"os"
)

type GameObjectType int

const (
	_ GameObjectType = iota
	GameObjMain
	GameObjShield
	GameObjBullet
	GameObjHommingBullet
	GameObjSuperBullet
	GameObjDeco
	GameObjMark
	GameObjEnd
)

type ClientActionType int

const (
	ActionAccel ClientActionType = iota
	ActionBullet
	ActionSuperBullet
	ActionHommingBullet
	ActionBurstBullet
	ActionEnd
)

type GameConfig struct {
	TcpListen            string
	WsListen             string
	ClearY               bool
	FramePerSec          float64
	AICountPerWorld      int
	AINames              []string
	MaxTcpClientPerWorld int
	MaxWsClientPerWorld  int
	StartWorldCount      int
	RemoveEmptyWorld     bool
	TcpClientEncode      string // gob , json
	WorldCube            *HyperRect
	WorldDiag            float64
	APIncFrame           int
	KillScore            int
	ShieldCount          int
	MaxObjectRadius      float64

	MoveLimit  [GameObjEnd]float64
	Radius     [GameObjEnd]float64
	ObjSqd     [GameObjEnd][GameObjEnd]float64
	IsInteract [GameObjEnd][GameObjEnd]bool // harmed obj : can harm obj
	AP         [ActionEnd]int
}

func ValidateConfig(config *GameConfig) {
	config.WorldDiag = config.WorldCube.DiagLen()
	for _, o := range config.Radius {
		if o > config.MaxObjectRadius {
			config.MaxObjectRadius = o
		}
	}
	for o1 := GameObjMain; o1 < GameObjEnd; o1++ {
		for o2 := GameObjMain; o2 < GameObjEnd; o2++ {
			config.ObjSqd[o1][o2] = math.Pow(config.Radius[o1]+config.Radius[o2], 2)
		}
	}

}

const WorldSize = 500

var defaultConfig = GameConfig{
	TcpListen:            "0.0.0.0:6666",
	WsListen:             "0.0.0.0:8080",
	FramePerSec:          60.0,
	RemoveEmptyWorld:     false,
	TcpClientEncode:      "gob",
	MaxTcpClientPerWorld: 8,
	MaxWsClientPerWorld:  8,
	StartWorldCount:      1,
	AICountPerWorld:      32,
	ClearY:               false,
	AINames:              []string{"AICloud", "AIRandom", "AI2", "AI3"},
	APIncFrame:           10,
	KillScore:            1,
	ShieldCount:          8,
	MaxObjectRadius:      1, // changed by init
	WorldCube: &HyperRect{
		Vector3D{-WorldSize, -WorldSize, -WorldSize},
		Vector3D{WorldSize, WorldSize, WorldSize}},

	MoveLimit: [GameObjEnd]float64{
		GameObjMain:          100,
		GameObjShield:        200,
		GameObjBullet:        300,
		GameObjHommingBullet: 200,
		GameObjSuperBullet:   600,
		GameObjDeco:          600,
		GameObjMark:          100},
	Radius: [GameObjEnd]float64{
		GameObjMain:          10,
		GameObjShield:        5,
		GameObjBullet:        5,
		GameObjHommingBullet: 7,
		GameObjSuperBullet:   15,
		GameObjDeco:          3,
		GameObjMark:          3},
	IsInteract: [GameObjEnd][GameObjEnd]bool{
		GameObjMain: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true},
		GameObjShield: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true},
		GameObjBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true},
		GameObjHommingBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        false,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true},
		GameObjSuperBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        false,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true},
	},
	AP: [ActionEnd]int{
		ActionAccel:         1,
		ActionBullet:        10,
		ActionSuperBullet:   80,
		ActionHommingBullet: 100,
		ActionBurstBullet:   10,
	},
}

var GameConst = defaultConfig

func SaveConfig(config *GameConfig, filename string) bool {
	j, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Printf("err in make json %v", err)
		return false
	}
	fd, err := os.Create(filename)
	if err != nil {
		log.Printf("err in create %v", err)
		return false
	}
	defer fd.Close()
	n, err := fd.Write(j)
	if err != nil {
		log.Printf("err in write %v %v", n, err)
		return false
	}
	return true
}

func LoadConfig(config *GameConfig, filename string) bool {
	fd, err := os.Open(filename)
	if err != nil {
		log.Printf("err in open %v", err)
		return false
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)
	err = dec.Decode(config)
	if err != nil {
		log.Printf("err in decode %v ", err)
		return false
	}
	ValidateConfig(config)
	return true
}

func SaveLoad(config *GameConfig, filename string) {
	ValidateConfig(config)
	SaveConfig(config, filename)
	LoadConfig(config, filename)
	log.Printf("%v", config)
}

func init() {
	ValidateConfig(&GameConst)
	//SaveLoad(&defaultConfig, "default.json")
}
