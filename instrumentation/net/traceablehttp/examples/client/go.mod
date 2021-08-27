module github.com/traceableai/goagent/instrumentation/net/traceablehttp/examples/client

go 1.16

require (
	github.com/traceableai/goagent v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.18.0
)

replace github.com/traceableai/goagent => ../../../../../
