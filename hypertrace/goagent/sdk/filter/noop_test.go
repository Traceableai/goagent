package filter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopFilter(t *testing.T) {
	f := NoopFilter{}
	res := f.Evaluate(context.Background(), nil)
	assert.False(t, res.Block)
}
