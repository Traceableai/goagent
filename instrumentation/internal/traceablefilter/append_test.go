package traceablefilter

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk/filter"
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
			Opa: &traceableconfig.Opa{ // needed to run the test
				Endpoint:          traceableconfig.String("http//localhost:123"),
				PollPeriodSeconds: traceableconfig.Int32(10),
			},
		}}
	f, closer := appendTraceableFilterPerConfig(enabledConfig, zap.NewNop(), filter.NoopFilter{})
	defer closer()

	assert.IsType(t, &filter.MultiFilter{}, f)
}
