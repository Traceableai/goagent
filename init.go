package goagent // import "github.com/Traceableai/goagent"

import (
	"github.com/Traceableai/goagent/config"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/hypertrace/goagent/instrumentation/hypertrace"
)

// Init initializes Traceable tracing and returns a shutdown function to flush data immediately
// on a termination signal.
func Init(cfg *config.AgentConfig) func() {
	internalconfig.InitConfig(cfg.Blocking)
	return hypertrace.Init(cfg.Tracing)
}
