package otel // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp/otel"

import (
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/net/hyperhttp"
	sdkhttp "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/net/http"
	"github.com/Traceableai/goagent/instrumentation/internal/traceablefilter"
)

func WrapHandler(delegate http.Handler, opts ...Option) http.Handler {
	o := &sdkhttp.Options{}
	for _, opt := range opts {
		opt(o)
	}
	o.Filter = traceablefilter.AppendTraceableFilter(o.Filter)

	return hyperhttp.WrapHandler(delegate, o)
}
