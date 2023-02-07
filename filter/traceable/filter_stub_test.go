//go:build !linux || !traceable_filter
// +build !linux !traceable_filter

package traceable

// To verify the Mac OS X blocking stub.

import (
	"testing"
	"time"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type noopAttributes struct{}

func (l noopAttributes) GetValue(key string) interface{} {
	return nil
}

var _ sdk.AttributeList = noopAttributes{}

type noopSpan struct{}

var _ sdk.Span = noopSpan{}

func (s noopSpan) GetAttributes() sdk.AttributeList {
	return &noopAttributes{}
}

func (s noopSpan) SetAttribute(_ string, _ interface{}) {}

func (s noopSpan) SetError(_ error) {}

func (s noopSpan) SetStatus(_ sdk.Code, _ string) {}

func (s noopSpan) IsNoop() bool { return true }

func (s noopSpan) AddEvent(_ string, _ time.Time, _ map[string]interface{}) {}

func TestBlockingStub(t *testing.T) {
	f := NewFilter("", &traceableconfig.AgentConfig{}, zap.NewNop())
	assert.IsType(t, Filter{}, *f)
	assert.True(t, f.Start())
	filterResult := f.EvaluateURLAndHeaders(noopSpan{}, "", map[string][]string{})
	assert.False(t, filterResult.Block)
	filterResult = f.EvaluateBody(noopSpan{}, []byte{}, nil)
	assert.False(t, filterResult.Block)
	filterResult = f.Evaluate(noopSpan{}, "", []byte{}, nil)
	assert.False(t, filterResult.Block)
}
