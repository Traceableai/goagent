//go:build !linux
// +build !linux

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk/filter"
)

// NewFilter TODO
func NewFilter(config *traceableconfig.AgentConfig) filter.Filter {
	// TODO replace with libtraceable filter impl
	return filter.NoopFilter{}
}
