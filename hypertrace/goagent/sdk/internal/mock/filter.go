package mock

import (
	"context"

	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
)

type Filter struct {
	Evaluator func(aa sdk.AttributeAccessor) result.FilterResult
}

func (f Filter) Evaluate(_ context.Context, aa sdk.AttributeAccessor) result.FilterResult {
	if f.Evaluator == nil {
		return result.FilterResult{}
	}
	return f.Evaluator(aa)
}

func (f Filter) Stop() error {
	return nil
}
