package main

import (
	"encoding/gob"
	"encoding/json"
	"flag"
	"github.com/kasworld/go4game"
	"log"
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
	timer60Ch := time.Tick(time.Duration(1000/go4game.GameConst.FramePerSec) * time.Millisecond)
	//timer60Ch := time.Tick(1 * time.Microsecond)

	var dec go4game.IDecoder
	var enc go4game.IEncoder
	if go4game.GameConst.TcpClientEncode == "gob" {
		dec = gob.NewDecoder(conn)
		enc = gob.NewEncoder(conn)
	} else if go4game.GameConst.TcpClientEncode == "json" {
		dec = json.NewDecoder(conn)
		enc = json.NewEncoder(conn)
	} else {
		log.Fatal("unknown tcp client encode %v", go4game.GameConst.TcpClientEncode)
	}
	//log.Printf("client connected %v", conn)
clientloop:
	for {
		select {
		case <-timer60Ch:
			// sp := GamePacket{
			// 	Cmd: ReqWorldInfo,
			// }
			sp := go4game.ReqGamePacket{
				Cmd: go4game.ReqNearInfo,
			}
			err := enc.Encode(&sp)
			if err != nil {
				log.Printf("client %v", err)
				_ = err
				break clientloop
			}
			//log.Printf("%v\n", plen)
			var rp go4game.RspGamePacket
			err = dec.Decode(&rp)
			if err != nil {
				log.Printf("%v", err)
			}
			switch rp.Cmd {
			case go4game.RspNearInfo:
				//s, _ := json.MarshalIndent(rp.WorldInfo, "", "  ")
				//log.Printf("%v", string(s))
			case go4game.RspWorldInfo:
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

func main() {
	var connectTo = flag.String("connectTo", "localhost:6666", "run time sec")
	var rundur = flag.Int("rundur", 60, "run time sec")
	var gocount = flag.Int("client", 4, "client count")
	flag.Parse()
	log.Printf("client %v %v %v", *connectTo, *rundur, *gocount)

	ClientMain(*connectTo, *gocount, *rundur)
	time.Sleep(time.Duration(*rundur) * time.Second)

}
