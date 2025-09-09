package hypermux // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/github.com/gorilla/hypermux"

import (
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	sdkhttp "github.com/Traceableai/goagent/hypertrace/goagent/sdk/instrumentation/net/http"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func spanNameFormatter(operation string, r *http.Request) string {
	return getOperationNameFromRoute(r)
}

func getOperationNameFromRoute(r *http.Request) string {
	route := mux.CurrentRoute(r)
	spanName := ""
	if route != nil {
		var err error
		spanName, err = route.GetPathTemplate()
		if err != nil {
			spanName, _ = route.GetPathRegexp()
		}
	}

	if spanName == "" {
		// if somehow retrieving the path template or path regexp fails, we still
		// want to use the method as fallback.
		spanName = r.Method
	}
	return spanName
}

// NewMiddleware sets up a handler to start tracing the incoming requests.
func NewMiddleware(options *sdkhttp.Options) mux.MiddlewareFunc {
	mh := opentelemetry.NewHttpOperationMetricsHandler(getOperationNameFromRoute)
	return func(delegate http.Handler) http.Handler {
		return otelhttp.NewHandler(
			sdkhttp.WrapHandler(delegate, opentelemetry.SpanFromContext, options, map[string]string{}, mh),
			"",
			otelhttp.WithSpanNameFormatter(spanNameFormatter),
		)
	}
}
