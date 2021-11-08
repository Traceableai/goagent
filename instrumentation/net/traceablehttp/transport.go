package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import (
	"github.com/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"
)

// NewTransport wraps the provided http.RoundTripper with one that
// starts a span and injects the span context into the outbound request headers.
var NewTransport = hyperhttp.NewTransport
