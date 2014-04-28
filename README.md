go4game
=======

### game server framework using  golang

![web client screenshot](/Screenshot.png?raw=true)


    GameService : server.go : main service entry
        World : world.go : game world or zone, terrain
            Team : team.go : AI or user controlled object
                GameObject : gameobject.go : component of team
                ConnInfo : common.go : tcp connection to channel
                    PacketStat : packet statistics


### requirement for websocket, web client

for 3d web

- threejs from threejs.org

- ( included in www/js folder )

for websocket

- https://github.com/gorilla/websocket

- ( clone yourself )


### execute
execute server

    go run server/main.go -rundur 60

execute client

    go run client/main.go -client 1000 -rundur 60 -connectTo localhost:6666

exec web client

    http://localhost:8080/www/client3d.html

### korean discription

http://kasw.blogspot.kr/2014/04/go4game.html

http://kasw.blogspot.kr/2014/04/go4game-go.html

주요 object 정의

Server : serverMain 과 같은 것, 실행된 서버 instance , 여러개의 world를 관장한다.

World : World : ai와 user가 접속해서 interaction하는 공간. 실 game contents가 일어나는 곳.

Team : gameObject list , AI , user 경쟁의 단위

GameObject : team을 구성하는 object

ConnInfo : net.Conn to channel, 2 goroutine

PacketStat : packet 통계용

Cmd : cmd channel object

