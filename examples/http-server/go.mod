module github.com/Traceableai/goagent/examples/http-server

go 1.16

require (
	github.com/Traceableai/goagent v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.21.0
)

replace github.com/Traceableai/goagent => ../../

replace github.com/Traceableai/agent-config/gen/go => ../../../agent-config/gen/go
