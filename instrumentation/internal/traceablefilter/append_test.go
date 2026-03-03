package traceablefilter

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestAppendTraceableFilter(t *testing.T) {
	enabledConfig := &config.AgentConfig{
		Tracing: &hyperconfig.AgentConfig{
			ServiceName: hyperconfig.String("test-service"),
		},
		TraceableConfig: &traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(true),
			},
			DetectionConfig: &traceableconfig.ThreatActivityDetection{
				Enabled: traceableconfig.Bool(false),
			},
		}}
	f, closer := appendTraceableFilterPerConfig(enabledConfig, zap.NewNop(), filter.NoopFilter{})
	defer closer()

	assert.IsType(t, &filter.MultiFilter{}, f)
}
