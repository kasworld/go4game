go4game
=======

game server framework using  golang

    GameService : server.go 
        World : world.go 
            Team : team.go 
                GameObject : gameobject.go 


execute server and client 
go run runserver/main.go -client 1000


### korean discription

주요 object 정의

Server : serverMain 과 같은 것, 실행된 서버 instance , 여러개의 world를 관장한다.

World : World : ai와 user가 접속해서 interaction하는 공간. 실 game contents가 일어나는 곳.

Team : gameObject list , AI , user 경쟁의 단위

GameObject : team을 구성하는 object 
