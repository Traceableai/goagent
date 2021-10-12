module github.com/Traceableai/goagent

go 1.15

require (
	github.com/Traceableai/agent-config/gen/go v0.0.0-20210928164718-dbda6f7a1ec6
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/hypertrace/agent-config/gen/go v0.0.0-20210928174043-f66d46bdfecb
	github.com/hypertrace/goagent v0.4.1-0.20211010180325-48385d780501
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.25.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.25.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.0.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.1 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.0.1 // indirect
	go.opentelemetry.io/otel/sdk v1.0.1
	go.opentelemetry.io/otel/trace v1.0.1
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20211008194852-3b03d305991f // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211008145708-270636b82663 // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)
