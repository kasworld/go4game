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
	var gocount = flag.Int("gocount", 1, "goroutine count")
	flag.Parse()
	log.Printf("%v %v %v", connectTo, *rundur, *gocount)

	service := *go4game.NewGameService(connectTo)
	service.CmdCh <- go4game.Cmd{Cmd: "start"}
	time.Sleep(1 * time.Second)
	go4game.ClientMain(connectTo, *gocount)
	time.Sleep(time.Duration(*rundur) * time.Second)
}
