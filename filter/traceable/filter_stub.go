//go:build !linux || !traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	"context"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
	"go.uber.org/zap"
)

// NewFilter creates libtraceable based blocking filter.
// It takes agent config and logger as parameters for creating a corresponding filter.
func NewFilter(_ *traceableconfig.AgentConfig, l *zap.Logger) *Filter {
	l.Debug("Using NOOP traceable filter.")
	return &Filter{}
}

type Filter struct{}

var _ filter.Filter = (*Filter)(nil)

// Start() starts the threads to poll config
func (f Filter) Start() bool { return true }

// Evaluate calls into libtraceable to evaluate if request should be blocked
func (Filter) Evaluate(context.Context, sdk.AttributeAccessor) result.FilterResult {
	return result.FilterResult{}
}

func (Filter) Stop() error {
	return nil
}
