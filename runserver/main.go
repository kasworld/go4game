package main

import (
	"flag"
	"fmt"
	"github.com/kasworld/go4game"
	"time"
)

func main() {
	connectTo := "localhost:6666"
	var rundur = flag.Int("rundur", 60, "run time sec")
	var gocount = flag.Int("gocount", 4, "goroutine count")
	flag.Parse()
	fmt.Printf("%v %v %v\n", connectTo, *rundur, *gocount)

	service := *go4game.NewGameService(connectTo)
	service.CmdCh <- go4game.Cmd{Cmd: "start"}
	//go gameService.Run()
	time.Sleep(1 * time.Second)
	go4game.ClientMain(connectTo, *gocount)
	time.Sleep(time.Duration(*rundur) * time.Second)
}
