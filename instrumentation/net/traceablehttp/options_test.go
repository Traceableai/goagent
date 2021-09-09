package traceablehttp

import (
	"testing"

	"github.com/hypertrace/goagent/sdk/instrumentation/net/http"
	"github.com/stretchr/testify/assert"
)

func TestNilOptions(t *testing.T) {
	opts := (*Options)(nil)
	assert.Equal(t, http.Options{}, *opts.toSDKOptions())
}
