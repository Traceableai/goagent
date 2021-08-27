package goagent

import (
	"github.com/hypertrace/goagent/instrumentation/hypertrace"
	"github.com/traceableai/goagent/config"
	internalconfig "github.com/traceableai/goagent/internal/config"
)

// Init initializes hypertrace tracing and returns a shutdown function to flush data immediately
// on a termination signal.
func Init(cfg *config.AgentConfig) func() {
	internalconfig.InitConfig(&cfg.Traceable)
	return hypertrace.Init(cfg.Hypertrace)
}
