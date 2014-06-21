package snakebase

import (
	"encoding/json"
	//"fmt"
	"github.com/kasworld/go4game"
	"log"
	"os"
	//"runtime"
	//"time"
)

type SnakeConfig struct {
	WorldCube   *go4game.HyperRect
	WorldDiag   float64
	WorldCube2  *go4game.HyperRect // for octree
	WorldDiag2  float64
	FramePerSec float64
}

func (config *SnakeConfig) Validate() {
	config.WorldDiag = config.WorldCube.DiagLen()
	config.WorldCube2 = &go4game.HyperRect{
		Min: config.WorldCube.Min.Sub(go4game.Vector3D{100, 100, 100}),
		Max: config.WorldCube.Max.Add(go4game.Vector3D{100, 100, 100}),
	}
	config.WorldDiag2 = config.WorldCube2.DiagLen()
}
func (config *SnakeConfig) Save(filename string) bool {
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
func (config *SnakeConfig) Load(filename string) bool {
	return true
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
func (config *SnakeConfig) SaveLoad(filename string) {
	config.Validate()
	config.Save(filename)
	config.Load(filename)
	log.Printf("%v", config)
}
func (config *SnakeConfig) NewService() ServiceI {
	g := SnakeService{
		id:     <-go4game.IdGenCh,
		cmdCh:  make(chan go4game.GoCmd, 10),
		Worlds: make(map[int64]WorldI),
		config: config,
	}
	g.AddWorld(g.NewWorld())
	return &g
}

const WorldSize = 500
const WorldSizeY = 500

var GameConst = SnakeConfig{
	WorldCube: &go4game.HyperRect{
		go4game.Vector3D{-WorldSize, -WorldSizeY, -WorldSize},
		go4game.Vector3D{WorldSize, WorldSizeY, WorldSize}},
	FramePerSec: 60.0,
}
