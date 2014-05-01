go build server.go
./server -rundur 60 -pfilename server.pprof
go tool pprof server server.pprof
rm server server.pprof
