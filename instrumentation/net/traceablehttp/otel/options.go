package otel // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp/otel"

import (
	"github.com/hypertrace/goagent/sdk/filter"
	sdkhttp "github.com/hypertrace/goagent/sdk/instrumentation/net/http"
)

type Option func(o *sdkhttp.Options)

func WithFilter(f filter.Filter) Option {
	return func(o *sdkhttp.Options) {
		o.Filter = f
	}
}
