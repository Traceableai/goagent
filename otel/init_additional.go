package otel // import "github.com/Traceableai/goagent/otel"

import (
	"os"

	"github.com/Traceableai/goagent/config"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	hyperotel "github.com/hypertrace/goagent/instrumentation/opentelemetry"
	"go.opentelemetry.io/otel/sdk/trace"
)

// InitAsAdditional initializes opentelemetry tracing and returns a span processor and a shutdown
// function to flush data immediately on a termination signal.
// This is ideal for when we use goagent along with other opentelemetry setups.
func InitAsAdditional(cfg *config.AgentConfig) (trace.SpanProcessor, func()) {
	loggerCloser := logger.InitLogger(os.Getenv("TA_LOG_LEVEL"))
	internalstate.AppendCloser(loggerCloser)

	if cfg.Tracing.Enabled.Value {
		internalstate.InitConfig(cfg)
	} else {
		internalstate.InitConfig(internalconfig.DisabledConfig)
	}

	sp, tracingCloser := hyperotel.InitAsAdditional(cfg.Tracing)
	internalstate.AppendCloser(tracingCloser)
	return sp, internalstate.CloserFn()
}

var attrsRemovalPrefixes = []string{
	"http.request.header.",
	"http.response.header.",
	"http.request.body",
	"http.response.body",
	"rpc.request.metadata.",
	"rpc.response.metadata.",
	"rpc.request.body",
	"rpc.response.body",
	"traceableai.",
}

var RemoveGoAgentAttrs = hyperotel.MakeRemoveGoAgentAttrs(attrsRemovalPrefixes)
