//go:build !linux || !traceable_filter
// +build !linux !traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
	"go.uber.org/zap"
)

// NewFilter creates libtraceable based blocking filter
func NewFilter(_ *traceableconfig.AgentConfig, l *zap.Logger) *Filter {
	l.Debug("Using NOOP traceable filter.")
	return &Filter{}
}

type Filter struct{}

var _ filter.Filter = (*Filter)(nil)

// Start() starts the threads to poll config
func (f Filter) Start() bool { return true }

func (f Filter) Stop() bool { return true }

// EvaluateURLAndHeaders calls into libtraceable to evaluate if request with URL should be blocked
// or if request with headers should be blocked
func (Filter) EvaluateURLAndHeaders(_ sdk.Span, _ string, _ map[string][]string) bool {
	return false
}

// EvaluateBody calls into libtraceable to evaluate if request with body should be blocked
func (Filter) EvaluateBody(_ sdk.Span, _ []byte, _ map[string][]string) bool {
	return false
}
