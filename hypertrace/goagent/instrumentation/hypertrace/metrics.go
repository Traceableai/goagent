package hypertrace // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace"

import (
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
)

var NewHttpOperationMetricsHandler = opentelemetry.NewHttpOperationMetricsHandler
