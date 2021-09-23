//go:build !linux
// +build !linux

package traceable

// To verify the Mac OS X blocking stub.

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/stretchr/testify/assert"
)

type noopSpan struct{}

var _ sdk.Span = noopSpan{}

func (s noopSpan) SetAttribute(_ string, _ interface{}) {}

func (s noopSpan) SetError(_ error) {}

func (s noopSpan) SetStatus(_ sdk.Code, _ string) {}

func (s noopSpan) IsNoop() bool { return true }

func TestBlockingStub(t *testing.T) {
	f := NewFilter(&traceableconfig.AgentConfig{})
	assert.IsType(t, Filter{}, *f)
	assert.True(t, f.Start())
	assert.False(t, f.EvaluateURLAndHeaders(noopSpan{}, "", map[string][]string{}))
	assert.False(t, f.EvaluateBody(noopSpan{}, []byte{}))
}
