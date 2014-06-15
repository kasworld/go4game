package go4game

import (
	"log"
	//"time"
	"encoding/json"
	"math"
	"os"
)

type GameConfig struct {
	TcpListen            string
	WsListen             string
	FramePerSec          float64
	AICountPerWorld      int
	AINames              []string
	MaxTcpClientPerWorld int
	MaxWsClientPerWorld  int
	StartWorldCount      int
	RemoveEmptyWorld     bool
	SetTerrain           bool
	SetFood              bool
	TcpClientEncode      string // gob , json
	WorldCube            *HyperRect
	WorldDiag            float64
	WorldCube2           *HyperRect // for octree
	WorldDiag2           float64
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

func (config *GameConfig) Validate() {
	config.WorldDiag = config.WorldCube.DiagLen()
	config.WorldCube2 = &HyperRect{
		Min: config.WorldCube.Min.Sub(Vector3D{100, 100, 100}),
		Max: config.WorldCube.Max.Add(Vector3D{100, 100, 100}),
	}
	config.WorldDiag2 = config.WorldCube2.DiagLen()

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

func (config *GameConfig) Save(filename string) bool {
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

func (config *GameConfig) Load(filename string) bool {
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
	config.Validate()
	return true
}

func (config *GameConfig) SaveLoad(filename string) {
	config.Validate()
	config.Save(filename)
	config.Load(filename)
	log.Printf("%v", config)
}

func init() {
	GameConst.Validate()
	//defaultConfig.SaveLoad("default.json")
}
