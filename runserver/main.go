package main

import (
	"flag"
	"github.com/kasworld/go4game"
	"log"
	"time"
)

func main() {
	connectTo := "localhost:6666"
	var rundur = flag.Int("rundur", 60, "run time sec")
	var client = flag.Int("client", 4, "client count")
	flag.Parse()
	log.Printf("%v %v %v", connectTo, *rundur, *client)

	service := *go4game.NewGameService(connectTo)
	service.CmdCh <- go4game.Cmd{Cmd: "start"}
	time.Sleep(1 * time.Second)
	go4game.ClientMain(connectTo, *client)
	time.Sleep(time.Duration(*rundur) * time.Second)
	service.CmdCh <- go4game.Cmd{Cmd: "quit"}
	time.Sleep(5 * time.Second)
}
