package traceablehttp

import (
	"net/http"

	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/filters/blocking"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry"
	"github.com/hypertrace/goagent/sdk/filter"
	sdkhttp "github.com/hypertrace/goagent/sdk/instrumentation/net/http"
)

// WrapHandler returns a new round tripper instrumented that relies on the
// needs to be used with OTel instrumentation.
func WrapHandler(delegate http.Handler, options *Options) http.Handler {
	newOpts := Options{
		Filter: resolveFilter(internalconfig.GetConfig(), options.Filter),
	}

	return sdkhttp.WrapHandler(
		delegate,
		opentelemetry.SpanFromContext,
		newOpts.toSDKOptions(),
	)
}

// isNoop returns true if the filter is NO-OP. This is useful specially when
// we are in environments where the filters can be noop se we can reduce the
// overhead of the filter call.
func isNoop(f filter.Filter) bool {
	if _, ok := f.(filter.NoopFilter); ok {
		return true
	}

	return false
}

func resolveFilter(cfg *traceconfig.AgentConfig, f filter.Filter) filter.Filter {
	blockingFilter := blocking.NewBlockingFilter(cfg)

	if !isNoop(blockingFilter) {
		if f != nil {
			return filter.NewMultiFilter(f)
		} else {
			return blockingFilter
		}
	}

	return f
}
