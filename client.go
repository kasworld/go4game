package go4game

import (
	"log"
	//"math/rand"
	"encoding/gob"
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

	var dec IDecoder
	var enc IEncoder
	if GameConst.TcpClientEncode == "gob" {
		dec = gob.NewDecoder(conn)
		enc = gob.NewEncoder(conn)
	} else if GameConst.TcpClientEncode == "json" {
		dec = json.NewDecoder(conn)
		enc = json.NewEncoder(conn)
	} else {
		log.Fatal("unknown tcp client encode %v", GameConst.TcpClientEncode)
	}

clientloop:
	for {
		select {
		case <-timer60Ch:
			// sp := GamePacket{
			// 	Cmd: ReqWorldInfo,
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
			case RspWorldInfo:
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
