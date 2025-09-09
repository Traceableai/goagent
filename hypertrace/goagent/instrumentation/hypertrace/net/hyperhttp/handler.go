package hyperhttp // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	sdkhttp "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/net/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewHandler wraps the passed handler, functioning like middleware.
func NewHandler(base http.Handler, operation string, opts ...Option) http.Handler {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	mh := opentelemetry.NewHttpOperationMetricsHandler(func(_ *http.Request) string { return operation })

	return otelhttp.NewHandler(
		sdkhttp.WrapHandler(base, opentelemetry.SpanFromContext, o.toSDKOptions(), map[string]string{}, mh),
		operation,
	)
}
