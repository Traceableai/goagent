package goagent // import "github.com/Traceableai/goagent"

import (
	"os"

	"github.com/Traceableai/goagent/config"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry"
	"go.opentelemetry.io/otel/trace"
)

// Init initializes Traceable tracing and returns a shutdown function to flush data immediately
// on a termination signal.
func Init(cfg *config.AgentConfig) func() {
	loggerCloser := logger.InitLogger(os.Getenv("TA_LOG_LEVEL"))
	internalstate.AppendCloser(loggerCloser)
	if cfg.Tracing.Enabled.Value {
		internalstate.InitConfig(cfg)
	} else {
		internalstate.InitConfig(internalconfig.DisabledConfig)
	}

	tracingCloser := opentelemetry.InitWithSpanProcessorWrapper(cfg.Tracing, &traceableSpanProcessorWrapper{})
	internalstate.AppendCloser(tracingCloser)
	return internalstate.CloserFn()
}

func RegisterService(
	serviceName string,
	resourceAttributes map[string]string,
) (SpanStarter, trace.TracerProvider, error) {
	s, tp, err := opentelemetry.RegisterServiceWithSpanProcessorWrapper(serviceName, resourceAttributes, &traceableSpanProcessorWrapper{})
	if err != nil {
		return nil, tp, err
	}

	return translateSpanStarter(s), tp, nil
}
