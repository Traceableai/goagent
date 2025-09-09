package sdk // import "github.com/Traceableai/goagent/hypertrace/goagent/sdk"

import (
	"net/http"
)

type HttpOperationMetricsHandler interface {
	AddToRequestCount(int64, *http.Request)
}
