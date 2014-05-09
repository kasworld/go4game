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
	flag.Parse()
	log.Printf("rundur:%vs profile:%v", *rundur, *profilefilename)

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
	time.Sleep(2 * time.Second)
}
