package mock

import (
	"context"

	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
)

type Filter struct {
	Evaluator func(span sdk.Span) result.FilterResult
}

func (f Filter) Evaluate(_ context.Context, span sdk.Span) result.FilterResult {
	if f.Evaluator == nil {
		return result.FilterResult{}
	}
	return f.Evaluator(span)
}

func (f Filter) Stop() error {
	return nil
}
