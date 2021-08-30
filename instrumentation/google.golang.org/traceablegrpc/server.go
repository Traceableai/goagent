package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/Traceableai/goagent/instrumentation/internal/filter"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry/google.golang.org/hypergrpc"
	"google.golang.org/grpc"
)

// WrapUnaryServerInterceptor returns a new unary server interceptor that will
// complement existing OpenTelemetry instrumentation
func WrapUnaryServerInterceptor(delegate grpc.UnaryServerInterceptor, options *Options) grpc.UnaryServerInterceptor {
	newOpts := Options{
		Filter: filter.ResolveFilter(internalconfig.GetConfig(), options.Filter),
	}

	return hypergrpc.WrapUnaryServerInterceptor(delegate, newOpts.toSDKOptions())

}
