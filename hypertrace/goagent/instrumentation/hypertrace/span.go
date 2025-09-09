package hypertrace // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace"

import (
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
)

type Span opentelemetry.Span

var (
	SpanFromContext = opentelemetry.SpanFromContext
	StartSpan       = opentelemetry.StartSpan
	NoopStartSpan   = opentelemetry.NoopStartSpan
)
