module github.com/Traceableai/goagent

go 1.15

require (
	github.com/Traceableai/agent-config/gen/go v0.0.0-20220804040331-6fb575c0338e
	github.com/gin-gonic/gin v1.8.1
	github.com/gorilla/mux v1.8.0
	github.com/hypertrace/agent-config/gen/go v0.0.0-20230126205246-bd4d81e696a6
	github.com/hypertrace/goagent v0.12.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.35.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.35.0
	go.opentelemetry.io/contrib/propagators/b3 v1.10.0
	go.opentelemetry.io/otel v1.10.0
	go.opentelemetry.io/otel/sdk v1.10.0
	go.opentelemetry.io/otel/trace v1.10.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
)
