package hypermux // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace/github.com/gorilla/hypermux"

import (
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/github.com/gorilla/hypermux"
	"github.com/gorilla/mux"
)

func NewMiddleware(opts ...Option) mux.MiddlewareFunc {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	return hypermux.NewMiddleware(o.toSDKOptions())
}
