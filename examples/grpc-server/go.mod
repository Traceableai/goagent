module github.com/Traceableai/goagent/examples/grpc-server

go 1.16

replace github.com/Traceableai/goagent => ../../

require (
	github.com/Traceableai/agent-config/gen/go v0.0.0-20210830210951-37afd5fed0fa // indirect
	github.com/Traceableai/goagent v0.0.0-00010101000000-000000000000
	github.com/hypertrace/goagent v0.3.1-0.20210827201008-0ff22ae72e11
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	google.golang.org/grpc v1.40.0
)
