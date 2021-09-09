package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/hypertrace/goagent/sdk/filter"
	"github.com/hypertrace/goagent/sdk/instrumentation/google.golang.org/grpc"
)

type Options struct {
	Filter filter.Filter
}

func (o *Options) toSDKOptions() *grpc.Options {
	if o == nil {
		return &grpc.Options{}
	}

	return &grpc.Options{
		Filter: o.Filter,
	}
}
