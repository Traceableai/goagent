module github.com/Traceableai/goagent/examples/http-server

go 1.16

require (
	github.com/Traceableai/goagent v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0
)

replace github.com/Traceableai/goagent => ../../
