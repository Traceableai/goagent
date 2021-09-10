package traceablegrpc // import "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"

import (
	"github.com/hypertrace/goagent/instrumentation/hypertrace/google.golang.org/hypergrpc"
	"github.com/hypertrace/goagent/sdk/filter"
)

type options struct {
	Filter filter.Filter
}

func (o *options) toHyperOptions() []hypergrpc.Option {
	if o == nil {
		return nil
	}

	opts := []hypergrpc.Option{}
	if o.Filter != nil {
		opts = append(opts, hypergrpc.WithFilter(o.Filter))
	}

	return opts
}

type Option func(o *options)

func WithFilter(f filter.Filter) Option {
	return func(o *options) {
		o.Filter = f
	}
}
