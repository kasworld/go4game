package go4game

import (
	"log"
	//"math/rand"
	"net"
	"time"
)

func ClientMain(connectTo string, clientcount int, rundur int) {
	log.Println("client Starting")
	for i := 0; i < clientcount; i++ {
		// time.Duration(rand.Float64()*60)
		go repeatReq(connectTo, time.Duration(rundur))
		//time.Sleep(1 * time.Millisecond)
	}
}

func repeatReq(connectTo string, rundur time.Duration) {
	conn, err := net.Dial("tcp", connectTo)
	if err != nil {
		log.Printf("client %v", err)
		return
	}
	defer conn.Close()
	timerCh := time.After(rundur * time.Second)
	timer60Ch := time.Tick(1000 / 60 * time.Millisecond)
clientloop:
	for {
		select {
		case <-timer60Ch:
			sendPacket := GamePacket{
				Cmd:       "makeTeam",
				Teamname:  "aaa",
				Teamcolor: []int{1, 2, 3},
			}
			_, err = writeJson(conn, &sendPacket)
			if err != nil {
				//log.Printf("client %v", err)
				_ = err
				break clientloop
			}

			var p GamePacket
			_, err := readJson(conn, &p)
			if err != nil {
				//log.Printf("client %v", err)
				_ = err
				break clientloop
			}
			//log.Println(p)
			_ = p
		case <-timerCh:
			break clientloop
		}
	}
	//log.Printf("client conn exit\n")
}
