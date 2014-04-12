go4game
=======

game server framework using  golang


    serverMain : start point
        World : client connect to
            NPCTeam : AI action
                gameObject : move , display
            UserTeam : client : user action ( or user simulated ai action )
                gameObject : move , display


    GameService
        World
            Team
                GameObject

### korean discription

주요 object 정의

Server : serverMain 과 같은 것, 실행된 서버 instance , 여러개의 world를 관장한다.

World : World : ai와 user가 접속해서 interaction하는 공간. 실 game contents가 일어나는 곳.

Team : gameObject list , AI , user 경쟁의 단위

AI team : ai 가 조종하는 server side의 team == NPC

user(client) team : user 또는 user simulateAI 가 조종하는 client side의 team

anyway live reload?