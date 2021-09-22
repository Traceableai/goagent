package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/instrumentation/internal/traceablefilter"

	"github.com/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"
)

// NewHandler wraps the passed handler, functioning like middleware.
func NewHandler(base http.Handler, operation string, opts ...Option) http.Handler {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	o.Filter = traceablefilter.AppendTraceableFilter(o.Filter)

	return hyperhttp.NewHandler(
		base,
		operation,
		o.toHyperOptions()...,
	)
}
