package filter

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/filter/traceable"
	sdkfilter "github.com/hypertrace/goagent/sdk/filter"
)

// isNoop returns true if the filter is NO-OP. This is useful specially when
// we are in environments where the filters can be noop se we can reduce the
// overhead of the filter call.
func isNoop(f sdkfilter.Filter) bool {
	_, isNoop := f.(sdkfilter.NoopFilter)
	return isNoop
}

// ResolveFilter resolves a joint filter based on the agent configuration and a provided filter.
// If both are nil or noop, this function will return nil.
func ResolveFilter(cfg *traceableconfig.AgentConfig, f sdkfilter.Filter) sdkfilter.Filter {
	blockingFilter := traceable.NewFilter(cfg)

	if isNoop(blockingFilter) {
		return f
	}

	if f != nil {
		return sdkfilter.NewMultiFilter(blockingFilter, f)
	}

	if isNoop(blockingFilter) {
		// if blockingFilter is also NoOp we return nil to avoid the overhead of
		// carrying a NoOp filter.
		return nil
	}

	return blockingFilter
}
