go4game
=======

# updated project https://github.com/kasworld/gowasm3dgame

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

see server info and index page

    http://localhost:8080/


exec web client or click from index page

    http://localhost:8080/www/client3d.html


### korean discription

http://kasw.blogspot.kr/2014/04/go4game.html

http://kasw.blogspot.kr/2014/04/go4game-go.html

http://kasw.blogspot.kr/2014/04/go4game.html

http://kasw.blogspot.kr/2014/05/go4game-ai.html

http://kasw.blogspot.kr/2014/05/go4game.html

http://kasw.blogspot.kr/2014/05/go4game_29.html
