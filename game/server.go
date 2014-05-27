package main

import (
	"flag"
	"github.com/kasworld/go4game"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	var rundur = flag.Int("rundur", 60*60*24*365, "run time sec")
	var profilefilename = flag.String("pfilename", "", "profile filename")
	var config = flag.String("config", "", "config filename")
	flag.Parse()
	log.Printf("rundur:%vs profile:%v config:%v",
		*rundur, *profilefilename, *config)

	if *config != "" {
		ok := go4game.LoadConfig(&go4game.GameConst, *config)
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
	service := *go4game.NewGameService()
	go service.Loop()
	service.CmdCh <- go4game.Cmd{Cmd: "start"}
	time.Sleep(time.Duration(*rundur) * time.Second)
	service.CmdCh <- go4game.Cmd{Cmd: "quit"}
	time.Sleep(1 * time.Second)
}
