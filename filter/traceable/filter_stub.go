//go:build !linux || !traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
	"github.com/hypertrace/goagent/sdk/filter/result"
	"go.uber.org/zap"
)

// NewFilter creates libtraceable based blocking filter.
// It takes tenant id, service name, agent config and logger as parameters for creating a corresponding filter.
// Library consumers which doesn't have access to tenant id should pass an empty string.
func NewFilter(_ string, _ string, _ *traceableconfig.AgentConfig, l *zap.Logger) *Filter {
	l.Debug("Using NOOP traceable filter.")
	return &Filter{}
}

type Filter struct{}

var _ filter.Filter = (*Filter)(nil)

// Start() starts the threads to poll config
func (f Filter) Start() bool { return true }

func (f Filter) Stop() bool { return true }

// Evaluate calls into libtraceable to evaluate if request should be blocked
func (Filter) Evaluate(_ sdk.Span) result.FilterResult {
	return result.FilterResult{}
}
