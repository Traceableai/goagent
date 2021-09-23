package traceablefilter

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/filter/traceable"
	internalstate "github.com/Traceableai/goagent/internal/state"
	sdkfilter "github.com/hypertrace/goagent/sdk/filter"
)

// AppendTraceableFilter resolves a joint filter based on the agent configuration and a provided filter.
// If both are nil or noop, this function will return nil.
func AppendTraceableFilter(f sdkfilter.Filter) sdkfilter.Filter {
	cfg := internalstate.GetConfig()
	f, closer := appendTraceableFilterPerConfig(cfg, f)
	internalstate.AppendCloser(closer)
	return f
}

func appendTraceableFilterPerConfig(cfg *traceableconfig.AgentConfig, f sdkfilter.Filter) (sdkfilter.Filter, func()) {
	if cfg.BlockingConfig == nil ||
		cfg.BlockingConfig.Enabled == nil ||
		!cfg.BlockingConfig.Enabled.Value {
		return f, func() {}
	}

	traceableFilter := traceable.NewFilter(cfg)
	if !traceableFilter.Start() {
		return f, func() {}
	}
	closer := func() { traceableFilter.Stop() }

	if f != nil {
		return sdkfilter.NewMultiFilter(traceableFilter, f), closer
	}

	return traceableFilter, closer
}
