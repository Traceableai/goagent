package filter

import (
	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/filters/blocking"
	sdkfilter "github.com/hypertrace/goagent/sdk/filter"
)

// isNoop returns true if the filter is NO-OP. This is useful specially when
// we are in environments where the filters can be noop se we can reduce the
// overhead of the filter call.
func isNoop(f sdkfilter.Filter) bool {
	if _, ok := f.(sdkfilter.NoopFilter); ok {
		return true
	}

	return false
}

func ResolveFilter(cfg *traceconfig.AgentConfig, f sdkfilter.Filter) sdkfilter.Filter {
	blockingFilter := blocking.NewBlockingFilter(cfg)

	if isNoop(blockingFilter) {
		return f
	}

	if f == nil {
		return blockingFilter
	}

	return sdkfilter.NewMultiFilter(f)
}
