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
	connectTo := "0.0.0.0:6666"
	wsCconnectTo := "0.0.0.0:8080"
	var rundur = flag.Int("rundur", 3600, "run time sec")
	var profilefilename = flag.String("pfilename", "", "profile filename")
	flag.Parse()
	log.Printf("Listen:%v wsListen:%v rundur:%vs profile:%v",
		connectTo, wsCconnectTo, *rundur, *profilefilename)

	if *profilefilename != "" {
		f, err := os.Create(*profilefilename)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	config := go4game.ServiceConfig{
		TcpListen:            connectTo,
		WsListen:             wsCconnectTo,
		NpcCountPerWorld:     2,
		MaxTcpClientPerWorld: 32,
		MaxWsClientPerWorld:  32,
		StartWorldCount:      1,
	}
	service := *go4game.NewGameService(&config)
	go service.Loop()
	service.CmdCh <- go4game.Cmd{Cmd: "start"}
	time.Sleep(time.Duration(*rundur) * time.Second)
	service.CmdCh <- go4game.Cmd{Cmd: "quit"}
	time.Sleep(2 * time.Second)
}
