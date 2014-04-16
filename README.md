go4game
=======

game server framework using  golang

    GameService : server.go : main service entry
        World : world.go : game world or zone, terrain
            Team : team.go : AI or user controlled object
                GameObject : gameobject.go : component of team
                ConnInfo : common.go : tcp connection to channel
                    StatInfo : packet statistics


execute server
go run runserver/main.go -rundur 60

execute client
go run runclient/main.go -client 1000 -rundur 60 -connectTo localhost:6666


### korean discription

주요 object 정의

Server : serverMain 과 같은 것, 실행된 서버 instance , 여러개의 world를 관장한다.

World : World : ai와 user가 접속해서 interaction하는 공간. 실 game contents가 일어나는 곳.

Team : gameObject list , AI , user 경쟁의 단위

GameObject : team을 구성하는 object

ConnInfo : net.Conn to channel, 2 goroutine

StatInfo : packet 통계용

Cmd : cmd channel object

