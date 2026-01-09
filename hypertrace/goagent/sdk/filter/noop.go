package filter // import "github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"

import (
	"context"

	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
)

// NoopFilter is a filter that always evaluates to false
type NoopFilter struct{}

var _ Filter = NoopFilter{}

// Evaluate that always returns false
func (NoopFilter) Evaluate(context.Context, sdk.Span) result.FilterResult {
	return result.FilterResult{}
}

func (NoopFilter) Stop() error {
	return nil
}
