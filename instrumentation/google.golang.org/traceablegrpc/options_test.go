package traceablegrpc

import (
	"testing"

	"github.com/hypertrace/goagent/sdk/filter"
	"github.com/stretchr/testify/assert"
)

func TestNilToHyperOptions(t *testing.T) {
	o := (*options)(nil)
	assert.Len(t, o.translateOptions(), 0)
}

func TestToHyperOptions(t *testing.T) {
	o := options{}
	assert.Len(t, o.translateOptions(), 0)

	o = options{Filter: filter.NoopFilter{}}
	assert.Len(t, o.translateOptions(), 1)
}
