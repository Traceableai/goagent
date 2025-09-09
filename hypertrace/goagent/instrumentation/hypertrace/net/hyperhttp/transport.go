package hyperhttp // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	sdkhttp "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/net/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewTransport wraps the provided http.RoundTripper with one that
// starts a span and injects the span context into the outbound request headers.
func NewTransport(base http.RoundTripper, opts ...Option) http.RoundTripper {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	return otelhttp.NewTransport(
		sdkhttp.WrapTransport(base, opentelemetry.SpanFromContext, o.toSDKOptions(), map[string]string{}),
	)
}
