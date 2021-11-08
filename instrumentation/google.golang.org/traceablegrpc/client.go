package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/hypertrace/goagent/instrumentation/hypertrace/google.golang.org/hypergrpc"
)

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor suitable
// for use in a grpc.Dial call.
var UnaryClientInterceptor = hypergrpc.UnaryClientInterceptor
