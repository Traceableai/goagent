package goagent // import "github.com/Traceableai/goagent"

import (
	"context"

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

func RegisterService(
	serviceName string,
	resourceAttributes map[string]string,
) (func(ctx context.Context, name string, opts ...Option) (context.Context, Span, func()), error) {
	s, err := hypertrace.RegisterService(serviceName, resourceAttributes)
	if err != nil {
		return nil, err
	}

	return htStarterToSpanStarter(s), nil
}
