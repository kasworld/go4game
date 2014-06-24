package main

import (
	"flag"
	//"github.com/kasworld/go4game"
	"github.com/kasworld/go4game/snakebase"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func run_main(sc *snakebase.SnakeConfig) {
	var rundur = flag.Int("rundur", 3, "run time sec")
	var profilefilename = flag.String("pfilename", "", "profile filename")
	var config = flag.String("config", "", "config filename")
	flag.Parse()
	log.Printf("rundur:%vs profile:%v config:%v",
		*rundur, *profilefilename, *config)

	if *config != "" {
		ok := snakebase.SnakeDefault.Load(*config)
		if !ok {
			log.Fatal("config load fail")
		}
	}

	if *profilefilename != "" {
		f, err := os.Create(*profilefilename)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	service := snakebase.NewService(sc)
	time.Sleep(time.Duration(*rundur) * time.Second)
	service.SendGoCmd("quit", nil, nil)
	time.Sleep(1 * time.Second)
}

func main() {
	run_main(&snakebase.SnakeDefault)
}
