package go4game

import (
	"log"
	//"math/rand"
	"encoding/json"
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
	timer60Ch := time.Tick(time.Duration(1000/GameConst.FramePerSec) * time.Millisecond)
	//timer60Ch := time.Tick(1 * time.Microsecond)
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

clientloop:
	for {
		select {
		case <-timer60Ch:
			// sp := GamePacket{
			// 	Cmd: ReqAIAct,
			// }
			sp := GamePacket{
				Cmd: ReqFrameInfo,
			}
			err := enc.Encode(&sp)
			if err != nil {
				log.Printf("client %v", err)
				_ = err
				break clientloop
			}
			//log.Printf("%v\n", plen)
			var rp GamePacket
			err = dec.Decode(&rp)
			if err != nil {
				log.Printf("%v", err)
			}
			switch rp.Cmd {
			case RspFrameInfo:
				//s, _ := json.MarshalIndent(rp.WorldInfo, "", "  ")
				//log.Printf("%v", string(s))
			default:
				log.Printf("unknown packet %v", rp)
			}
		case <-timerCh:
			break clientloop
		}
	}
	//log.Printf("client conn exit\n")
}
