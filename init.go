package goagent // import "github.com/Traceableai/goagent"

import (
	"os"

	"github.com/Traceableai/goagent/config"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	"github.com/hypertrace/goagent/instrumentation/hypertrace"
)

// Init initializes Traceable tracing and returns a shutdown function to flush data immediately
// on a termination signal.
func Init(cfg *config.AgentConfig) func() {
	loggerCloser := logger.InitLogger(os.Getenv("TA_LOG_LEVEL"))
	internalstate.AppendCloser(loggerCloser)

	if cfg.Tracing.Enabled.Value {
		internalstate.InitConfig(cfg.Blocking)
	} else {
		internalstate.InitConfig(internalconfig.DisabledConfig)
	}

	tracingCloser := hypertrace.Init(cfg.Tracing)
	internalstate.AppendCloser(tracingCloser)
	return internalstate.CloserFn()
}

func RegisterService(
	serviceName string,
	resourceAttributes map[string]string,
) (SpanStarter, error) {
	s, err := hypertrace.RegisterService(serviceName, resourceAttributes)
	if err != nil {
		return nil, err
	}

	return translateSpanStarter(s), nil
}
