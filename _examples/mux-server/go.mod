module github.com/Traceableai/goagent/_examples/mux-server

go 1.15

replace github.com/Traceableai/goagent => ../../

require (
	github.com/Traceableai/goagent v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0
	github.com/hypertrace/goagent v0.4.1-0.20211001093754-e2cf9aacff05
)
