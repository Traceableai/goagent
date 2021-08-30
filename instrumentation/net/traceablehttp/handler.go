package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/instrumentation/internal/filter"

	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry"
	sdkhttp "github.com/hypertrace/goagent/sdk/instrumentation/net/http"
)

// WrapHandler returns a new round tripper instrumented that relies on the
// needs to be used with OTel instrumentation.
func WrapHandler(delegate http.Handler, options *Options) http.Handler {
	newOpts := Options{
		Filter: filter.ResolveFilter(internalconfig.GetConfig(), options.Filter),
	}

	return sdkhttp.WrapHandler(
		delegate,
		opentelemetry.SpanFromContext,
		newOpts.toSDKOptions(),
	)
}
