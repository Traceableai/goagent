package traceablehttp

import (
	"testing"

	"github.com/hypertrace/goagent/sdk/filter"
	"github.com/stretchr/testify/assert"
)

func TestNilToHyperOptions(t *testing.T) {
	o := (*options)(nil)
	assert.Len(t, o.toHyperOptions(), 0)
}

func TestToHyperOptions(t *testing.T) {
	o := options{}
	assert.Len(t, o.toHyperOptions(), 0)

	o = options{Filter: filter.NoopFilter{}}
	assert.Len(t, o.toHyperOptions(), 1)
}
