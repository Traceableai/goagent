package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import (
	"net/http"

	"github.com/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"
)

// NewTransport wraps the provided http.RoundTripper with one that
// starts a span and injects the span context into the outbound request headers.
func NewTransport(base http.RoundTripper) http.RoundTripper {
	return hyperhttp.NewTransport(base)
}
