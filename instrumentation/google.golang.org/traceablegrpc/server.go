package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/Traceableai/goagent/instrumentation/internal/filter"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/hypertrace/google.golang.org/hypergrpc"
	"google.golang.org/grpc"
)

// WrapUnaryServerInterceptor returns a new unary server interceptor that will
// complement existing OpenTelemetry instrumentation
func UnaryServerInterceptor(options *Options) grpc.UnaryServerInterceptor {
	newOpts := Options{}
	if options == nil || options.Filter == nil {
		newOpts.Filter = filter.ResolveFilter(internalconfig.GetConfig(), nil)
	} else {
		newOpts.Filter = filter.ResolveFilter(internalconfig.GetConfig(), options.Filter)
	}

	return hypergrpc.UnaryServerInterceptor(newOpts.toSDKOptions())

}
