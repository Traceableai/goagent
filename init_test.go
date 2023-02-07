package goagent

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	internalstate "github.com/Traceableai/goagent/internal/state"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"github.com/stretchr/testify/assert"
)

func TestInitAgentIsDisabled(t *testing.T) {
	shutdown := Init(&config.AgentConfig{
		Tracing: &hyperconfig.AgentConfig{
			Enabled: config.Bool(false),
		},
		TraceableConfig: &traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: config.Bool(true),
			},
		},
	})
	defer shutdown()

	assert.False(t, internalstate.GetConfig().TraceableConfig.BlockingConfig.Enabled.Value)
}
