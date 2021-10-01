package traceablemux // import "github.com/Traceableai/goagent/instrumentation/github.com/gorilla/traceablemux"

import (
	"github.com/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"
	"github.com/hypertrace/goagent/sdk/filter"
)

type options struct {
	Filter filter.Filter
}

func (o *options) toHyperOptions() []hyperhttp.Option {
	if o == nil {
		return nil
	}

	opts := []hyperhttp.Option{}
	if o.Filter != nil {
		opts = append(opts, hyperhttp.WithFilter(o.Filter))
	}

	return opts
}

type Option func(o *options)

func WithFilter(f filter.Filter) Option {
	return func(o *options) {
		o.Filter = f
	}
}
