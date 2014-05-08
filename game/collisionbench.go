package main

import (
	"github.com/kasworld/go4game"
	"log"
	"time"
)

var colcount = 0

func IsCollision(sl go4game.SPObjList) bool {
	colcount += len(sl)
	return false
}

func main() {
	config := go4game.ServiceConfig{
		TcpListen:            "0.0.0.0:6666",
		WsListen:             "0.0.0.0:8080",
		NpcCountPerWorld:     1000,
		MaxTcpClientPerWorld: 32,
		MaxWsClientPerWorld:  32,
		StartWorldCount:      1,
	}
	service := *go4game.NewGameService(&config)
	for _, w := range service.Worlds {
		for i := 0; i < 10; i++ {
			w.Do1Frame(time.Now())
		}

		spp := w.MakeSpatialPartition()
		osum := 0
		for _, t := range w.Teams {
			osum += len(t.GameObjs)
		}
		log.Printf("%v, objs:%v, spp:%v, %v", w, osum, spp.PartCount, spp.PartSize)

		colcount = 0
		ss := go4game.CountStat{}
		st := time.Now()
		for i := 0; i < 10; i++ {
			for _, t := range w.Teams {
				for _, o := range t.GameObjs {
					spp.GetCollisionList(IsCollision, o.PosVector, spp.MaxObjectRadius)
					//spp.ApplyCollisionAction4(IsCollision, o)
					ss.Inc()
				}
			}
		}
		log.Printf("collision4 %v , %v", ss.CalcLap(time.Now().Sub(st)), colcount)

	}

}
