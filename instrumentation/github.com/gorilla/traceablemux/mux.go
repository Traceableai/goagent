package traceablemux // import "github.com/Traceableai/goagent/instrumentation/github.com/gorilla/traceablemux"

import (
	"github.com/Traceableai/goagent/instrumentation/internal/traceablefilter"
	"github.com/gorilla/mux"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry/github.com/gorilla/hypermux"
)

// NewMiddleware sets up a handler to start tracing the incoming requests.
func NewMiddleware(opts ...Option) mux.MiddlewareFunc {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	o.Filter = traceablefilter.AppendTraceableFilter(o.Filter)

	return hypermux.NewMiddleware(o.toHyperOptions())
}
