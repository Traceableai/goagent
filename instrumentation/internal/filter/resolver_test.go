package filter

import (
	"testing"

	sdkfilter "github.com/hypertrace/goagent/sdk/filter"
	"github.com/stretchr/testify/assert"
)

func TestIsNoop(t *testing.T) {
	assert.False(t, isNoop(nil))
	assert.True(t, isNoop(sdkfilter.NoopFilter{}))
}
