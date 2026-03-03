package filter // import "github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"

import (
	"context"
	"errors"

	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
)

// MultiFilter encapsulates multiple filters
type MultiFilter struct {
	filters []Filter
}

var _ Filter = (*MultiFilter)(nil)

// NewMultiFilter creates a new MultiFilter
func NewMultiFilter(filter ...Filter) *MultiFilter {
	return &MultiFilter{filters: filter}
}

// Evaluate runs body evaluators for each filter until one returns true
func (m *MultiFilter) Evaluate(ctx context.Context, span sdk.AttributeAccessor) result.FilterResult {
	for _, f := range m.filters {
		filterResult := f.Evaluate(ctx, span)
		if filterResult.Block {
			return filterResult
		}
	}
	return result.FilterResult{}
}

func (m *MultiFilter) Stop() error {
	var err error
	for _, f := range m.filters {
		err = errors.Join(err, f.Stop())
	}

	return err
}
