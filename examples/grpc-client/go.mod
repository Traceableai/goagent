module github.com/Traceableai/goagent/examples/grpc-client

go 1.16

replace github.com/Traceableai/goagent => ../../

require (
	github.com/Traceableai/goagent v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	google.golang.org/grpc v1.40.0
)
