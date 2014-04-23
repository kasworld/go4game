package main

import (
	"flag"
	//"fmt"
	"github.com/kasworld/go4game"
	"log"
	"time"
)

func main() {
	var connectTo = flag.String("connectTo", "localhost:6666", "run time sec")
	var rundur = flag.Int("rundur", 60, "run time sec")
	var gocount = flag.Int("client", 4, "client count")
	flag.Parse()
	log.Printf("client %v %v %v", *connectTo, *rundur, *gocount)

	go4game.ClientMain(*connectTo, *gocount, *rundur)
	time.Sleep(time.Duration(*rundur) * time.Second)

}
