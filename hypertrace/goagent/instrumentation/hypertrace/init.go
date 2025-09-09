package hypertrace // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace"

import (
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
)

// Init initializes hypertrace tracing and returns a shutdown function to flush data immediately
// on a termination signal.
var Init = opentelemetry.Init

var RegisterService = opentelemetry.RegisterService

// NewZapCore returns a new [zapcore.Core] which exports the logs to the configured exporter
var NewZapCore = opentelemetry.NewZapCore
