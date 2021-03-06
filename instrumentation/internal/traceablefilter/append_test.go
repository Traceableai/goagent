package traceablefilter

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk/filter"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestAppendTraceableFilter(t *testing.T) {
	enabledConfig := &traceableconfig.AgentConfig{
		BlockingConfig: &traceableconfig.BlockingConfig{
			Enabled: traceableconfig.Bool(true),
		},
		Opa: &traceableconfig.Opa{ // needed to run the test
			Endpoint:          traceableconfig.String("http//localhost:123"),
			PollPeriodSeconds: traceableconfig.Int32(10),
		},
	}
	f, closer := appendTraceableFilterPerConfig(enabledConfig, zap.NewNop(), filter.NoopFilter{})
	defer closer()

	assert.IsType(t, &filter.MultiFilter{}, f)
}

func TestAppendTraceableFilterWithTraceableFilterDisabled(t *testing.T) {
	disabledConfigs := map[string]*traceableconfig.AgentConfig{
		"no blocking config": {},
		"empty blocking config": {
			BlockingConfig: &traceableconfig.BlockingConfig{},
		},
		"disabled": {
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(false),
			},
		},
	}

	for name, cfg := range disabledConfigs {
		t.Run(name, func(t *testing.T) {
			f, closer := appendTraceableFilterPerConfig(cfg, zap.NewNop(), filter.NoopFilter{})
			assert.IsType(t, filter.NoopFilter{}, f)
			closer()
		})
	}
}
