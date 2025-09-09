package hyperhttp // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/net/hyperhttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	sdkhttp "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/net/http"
)

// WrapTransport wraps an uninstrumented RoundTripper (e.g. http.DefaultTransport)
// and returns an instrumented RoundTripper that has to be used as base for the
// OTel's RoundTripper.
func WrapTransport(delegate http.RoundTripper, options *sdkhttp.Options) http.RoundTripper {
	return sdkhttp.WrapTransport(delegate, opentelemetry.SpanFromContext, options, map[string]string{})
}
