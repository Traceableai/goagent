package hyperhttp // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/net/hyperhttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	sdkhttp "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/net/http"
)

// WrapHandler returns a new round tripper instrumented that relies on the
// needs to be used with OTel instrumentation.
func WrapHandler(delegate http.Handler, options *sdkhttp.Options) http.Handler {
	mh := opentelemetry.NewHttpOperationMetricsHandler(func(_ *http.Request) string { return "" })
	return sdkhttp.WrapHandler(delegate, opentelemetry.SpanFromContext, options, map[string]string{}, mh)
}
