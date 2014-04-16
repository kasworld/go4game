package main

import (
	"flag"
	"github.com/kasworld/go4game"
	"log"
	"time"
)

func main() {
	connectTo := "0.0.0.0:6666"
	var rundur = flag.Int("rundur", 60, "run time sec")
	flag.Parse()
	log.Printf("%v %v", connectTo, *rundur)

	service := *go4game.NewGameService(connectTo)
	service.CmdCh <- go4game.Cmd{Cmd: "start"}
	time.Sleep(time.Duration(*rundur) * time.Second)
	service.CmdCh <- go4game.Cmd{Cmd: "quit"}
	time.Sleep(2 * time.Second)
}
