//go:build !linux
// +build !linux

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
)

// NewFilter creates libtraceable based blocking filter
func NewFilter(config *traceableconfig.AgentConfig) *Filter {
	return &Filter{}
}

type Filter struct{}

var _ filter.Filter = (*Filter)(nil)

// Start() starts the threads to poll config
func (f Filter) Start() bool { return true }

func (f Filter) Stop() bool { return true }

// EvaluateURLAndHeaders calls into libtraceable to evaluate if request with URL should be blocked
// or if request with headers should be blocked
func (f Filter) EvaluateURLAndHeaders(span sdk.Span, url string, headers map[string][]string) bool {
	return false
}

// EvaluateBody calls into libtraceable to evaluate if request with body should be blocked
func (f Filter) EvaluateBody(span sdk.Span, body []byte) bool {
	return false
}
