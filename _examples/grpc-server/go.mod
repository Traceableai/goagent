module github.com/Traceableai/goagent/_examples/grpc-server

go 1.15

replace github.com/Traceableai/goagent => ../../

require (
	github.com/Traceableai/goagent v0.0.0-00010101000000-000000000000
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	google.golang.org/grpc v1.40.0
)
