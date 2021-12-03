package traceablegin // import "github.com/Traceableai/goagent/instrumentation/github.com/gin-gonic/traceablegin"

import (
	"github.com/hypertrace/goagent/instrumentation/hypertrace/github.com/gin-gonic/hypergin"
	"github.com/hypertrace/goagent/sdk/filter"
)

type options struct {
	Filter filter.Filter
}

func (o *options) translateOptions() []hypergin.Option {
	if o == nil {
		return nil
	}

	opts := []hypergin.Option{}
	if o.Filter != nil {
		opts = append(opts, hypergin.WithFilter(o.Filter))
	}

	return opts
}

type Option func(o *options)

func WithFilter(f filter.Filter) Option {
	return func(o *options) {
		o.Filter = f
	}
}
