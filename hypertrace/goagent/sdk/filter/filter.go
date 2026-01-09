package filter // import "github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"

import (
	"context"

	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
)

// Filter evaluates whether request should be blocked, `true` blocks the request and `false` continues it.
type Filter interface {
	// Evaluate can be used to evaluate URL, headers and body content in one call
	Evaluate(context.Context, sdk.Span) result.FilterResult
	Stop() error
}
