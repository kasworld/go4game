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
	GameObjNil GameObjectType = iota
	GameObjMain
	GameObjShield
	GameObjBullet
	GameObjHommingBullet
	GameObjSuperBullet
	GameObjDeco
	GameObjMark
	GameObjHard
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
	MinObjectRadius      float64

	MoveLimit  [GameObjEnd]float64
	Radius     [GameObjEnd]float64
	ObjSqd     [GameObjEnd][GameObjEnd]float64
	IsInteract [GameObjEnd][GameObjEnd]bool // harmed obj : can harm obj
	NoInteract [GameObjEnd]bool
	AP         [ActionEnd]int
}

func ValidateConfig(config *GameConfig) {
	config.WorldDiag = config.WorldCube.DiagLen()
	for _, o := range config.Radius {
		if o > config.MaxObjectRadius {
			config.MaxObjectRadius = o
		}
		if o > 0 && config.MinObjectRadius == 0 {
			config.MinObjectRadius = o
		}
		if o > 0 && o < config.MinObjectRadius {
			config.MinObjectRadius = o
		}
	}
	for o1 := GameObjMain; o1 < GameObjEnd; o1++ {
		for o2 := GameObjMain; o2 < GameObjEnd; o2++ {
			config.ObjSqd[o1][o2] = math.Pow(config.Radius[o1]+config.Radius[o2], 2)
		}
	}
	for i := GameObjNil; i < GameObjEnd; i++ {
		interact := false
		for j := GameObjNil; j < GameObjEnd; j++ {
			if config.IsInteract[i][j] {
				interact = true
				break
			}
			if config.IsInteract[j][i] {
				interact = true
				break
			}
		}
		if interact == false {
			config.NoInteract[i] = true
		}
	}
}

const WorldSize = 500

var defaultConfig = GameConfig{
	AICountPerWorld: 9,
	ClearY:          true,
	StartWorldCount: 1,
	//AINames:            []string{"AINothing", "AICloud", "AIRandom", "AI2", "AI3"},
	//AINames:            []string{"AICloud", "AIRandom", "AI2", "AI3"},
	AINames:              []string{"AI2", "AI3", "AI4"},
	APIncFrame:           10,
	ShieldCount:          8,
	MaxTcpClientPerWorld: 32,
	MaxWsClientPerWorld:  32,
	RemoveEmptyWorld:     false,
	TcpClientEncode:      "gob",
	TcpListen:            "0.0.0.0:6666",
	WsListen:             "0.0.0.0:8080",
	FramePerSec:          60.0,
	KillScore:            1,
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
		GameObjMark:          3,
		GameObjHard:          10},
	IsInteract: [GameObjEnd][GameObjEnd]bool{
		GameObjMain: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjShield: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjHommingBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        false,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjSuperBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        false,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
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
	//SaveLoad(&defaultConfig, "profile.json")
}
