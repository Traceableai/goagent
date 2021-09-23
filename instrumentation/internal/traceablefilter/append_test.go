package traceablefilter

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk/filter"

	"github.com/stretchr/testify/assert"
)

func TestAppendTraceableFilter(t *testing.T) {
	enabledConfig := &traceableconfig.AgentConfig{
		BlockingConfig: &traceableconfig.BlockingConfig{
			Enabled: traceableconfig.Bool(true),
		},
		Opa: &traceableconfig.Opa{ // needed to run the test
			Endpoint:          traceableconfig.String("localhost:123"),
			PollPeriodSeconds: traceableconfig.Int32(10),
		},
	}
	f := appendTraceableFilterPerConfig(enabledConfig, filter.NoopFilter{})
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
			f := appendTraceableFilterPerConfig(cfg, filter.NoopFilter{})
			assert.IsType(t, filter.NoopFilter{}, f)
		})
	}
}
