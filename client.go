package go4game

import (
	"log"
	"math/rand"
	"net"
	"time"
)

func ClientMain(connectTo string, gocount int) {
	log.Println("Starting client")
	for i := 0; i < gocount; i++ {
		go repeatReq(connectTo)
		time.Sleep(1 * time.Millisecond)
	}
}

func repeatReq(connectTo string) {
	conn, err := net.Dial("tcp", connectTo)
	if err != nil {
		log.Print(err)
		return
	}
	defer conn.Close()
	timerCh := time.After(time.Duration(rand.Float64()*60) * time.Second)
	timer60Ch := time.Tick(1000 / 10 * time.Millisecond)
clientloop:
	for {
		select {
		case <-timerCh:
			break clientloop
		case <-timer60Ch:
			sendPacket := GamePacket{
				Cmd:       "makeTeam",
				Teamname:  "aaa",
				Teamcolor: []int{1, 2, 3},
			}
			_, err = writeJson(conn, &sendPacket)
			if err != nil {
				if err.Error() != "EOF" {
					log.Print(err)
				}
				break clientloop
			}

			var p GamePacket
			_, err := readJson(conn, &p)
			if err != nil {
				if err.Error() != "EOF" {
					log.Print(err)
				}
				break clientloop
			}
			//log.Println(p)
			_ = p
		}
	}
	//log.Printf("client conn exit\n")
}
