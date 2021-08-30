package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/hypertrace/goagent/instrumentation/opentelemetry/google.golang.org/hypergrpc"
)

// WrapUnaryClientInterceptor returns a new unary client interceptor that will
// complement existing OpenTelemetry instrumentation
var WrapUnaryClientInterceptor = hypergrpc.WrapUnaryClientInterceptor
