go build server/main.go
./main -rundur 60 -pfilename server.pprof
go tool pprof main server.pprof
rm main server.pprof
