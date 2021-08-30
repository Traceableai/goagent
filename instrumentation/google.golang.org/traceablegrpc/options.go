package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/hypertrace/goagent/sdk/filter"
	sdkgrpc "github.com/hypertrace/goagent/sdk/instrumentation/google.golang.org/grpc"
)

type Options struct {
	Filter filter.Filter
}

func (o Options) toSDKOptions() *sdkgrpc.Options {
	return &sdkgrpc.Options{
		Filter: o.Filter,
	}
}
