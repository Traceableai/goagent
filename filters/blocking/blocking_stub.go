//+build !linux

package blocking

import (
	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk/filter"
)

// NewBlockingFilter TODO
func NewBlockingFilter(config *traceconfig.AgentConfig) filter.Filter {
	// TODO replace with libtraceable filter impl
	return filter.NoopFilter{}
}
