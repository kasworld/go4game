go build server.go
./server -rundur 360 -pfilename server.pprof
go tool pprof server server.pprof
rm server server.pprof
