package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/Traceableai/goagent/instrumentation/internal/filter"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/hypertrace/google.golang.org/hypergrpc"
	"google.golang.org/grpc"
)

// WrapUnaryServerInterceptor returns a new unary server interceptor that will
// complement existing OpenTelemetry instrumentation
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	o.Filter = filter.ResolveFilter(internalconfig.GetConfig(), o.Filter)

	return hypergrpc.UnaryServerInterceptor(o.toHyperOptions()...)

}
