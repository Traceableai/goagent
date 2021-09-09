package traceablehttp // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"

import (
	"net/http"

	"github.com/Traceableai/goagent/instrumentation/internal/filter"

	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/hypertrace/net/hyperhttp"
)

// NewHandler wraps the passed handler, functioning like middleware.
func NewHandler(base http.Handler, operation string, options *Options) http.Handler {
	newOpts := Options{
		Filter: filter.ResolveFilter(internalconfig.GetConfig(), options.Filter),
	}

	return hyperhttp.NewHandler(
		base,
		operation,
		newOpts.toSDKOptions(),
	)
}
