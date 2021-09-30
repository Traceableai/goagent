package traceablefilter

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/filter/traceable"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	sdkfilter "github.com/hypertrace/goagent/sdk/filter"
	"go.uber.org/zap"
)

// AppendTraceableFilter resolves a joint filter based on the agent configuration and a provided filter.
// If both are nil or noop, this function will return nil.
func AppendTraceableFilter(f sdkfilter.Filter) sdkfilter.Filter {
	cfg := internalstate.GetConfig()
	l := logger.GetLogger()
	f, closer := appendTraceableFilterPerConfig(cfg, l, f)
	internalstate.AppendCloser(closer)
	return f
}

func appendTraceableFilterPerConfig(cfg *traceableconfig.AgentConfig, l *zap.Logger, f sdkfilter.Filter) (sdkfilter.Filter, func()) {
	if cfg.BlockingConfig == nil ||
		cfg.BlockingConfig.Enabled == nil ||
		!cfg.BlockingConfig.Enabled.Value {
		return f, func() {}
	}

	traceableFilter := traceable.NewFilter(cfg, l)
	if !traceableFilter.Start() {
		return f, func() {}
	}
	closer := func() { traceableFilter.Stop() }

	l.Debug("Traceable filter started successfully")

	if f != nil {
		return sdkfilter.NewMultiFilter(traceableFilter, f), closer
	}

	return traceableFilter, closer
}
