package traceablehttp

import (
	"net/http"

	"github.com/hypertrace/goagent/instrumentation/opentelemetry"
	"github.com/hypertrace/goagent/sdk/filter"
	sdkhttp "github.com/hypertrace/goagent/sdk/instrumentation/net/http"
	"github.com/traceableai/goagent/filters/blocking"
	internalconfig "github.com/traceableai/goagent/internal/config"
)

// WrapHandler returns a new round tripper instrumented that relies on the
// needs to be used with OTel instrumentation.
func WrapHandler(delegate http.Handler, options *Options) http.Handler {
	if options != nil && options.Filter != nil {
		cfg := internalconfig.GetConfig()
		options.Filter = filter.NewMultiFilter(options.Filter, blocking.NewBlockingFilter(cfg))
	}

	return sdkhttp.WrapHandler(delegate, opentelemetry.SpanFromContext, options.toSDKOptions())
}
