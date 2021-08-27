//+build !linux

package blocking

import (
	"github.com/hypertrace/goagent/sdk/filter"
	traceconfig "github.com/traceableai/agent-config/gen/go/v1"
)

// NewBlockingFilter TODO
func NewBlockingFilter(config *traceconfig.Traceable) filter.Filter {
	// TODO replace with libtraceable filter impl
	return filter.NoopFilter{}
}
