package traceablefilter

import (
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/filter/traceable"
	sdkfilter "github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
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

func appendTraceableFilterPerConfig(cfg *config.AgentConfig, l *zap.Logger, f sdkfilter.Filter) (sdkfilter.Filter, func()) {
	traceableFilter := traceable.NewFilter(cfg.TraceableConfig, l)
	if !traceableFilter.Start() {
		return f, func() {}
	}
	closer := func() {
		if err := traceableFilter.Stop(); err != nil {
			l.Error("Failed to stop traceable filter", zap.Error(err))
		}
	}

	l.Debug("Traceable filter appended successfully")
	if f != nil {
		return sdkfilter.NewMultiFilter(traceableFilter, f), closer
	}

	return traceableFilter, closer
}
