package config

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoadIsNotOverridenByDefaults(t *testing.T) {
	cfg := &AgentConfig{
		Tracing: &hyperconfig.AgentConfig{
			DataCapture: &hyperconfig.DataCapture{
				RpcMetadata: &hyperconfig.Message{
					Request: hyperconfig.Bool(false),
				},
			},
		},
		Blocking: &traceableconfig.AgentConfig{
			Opa: &traceableconfig.Opa{
				Enabled: traceableconfig.Bool(false),
			},
		},
	}

	assert.Equal(t, false, cfg.Tracing.DataCapture.RpcMetadata.Request.Value)
	assert.Equal(t, false, cfg.Blocking.Opa.Enabled.Value)

	LoadEnv(cfg)
	// we verify here the value isn't overridden by default value (true)
	assert.Equal(t, false, cfg.Tracing.DataCapture.RpcMetadata.Request.Value)
	// we verify default value is used for undefined value (true)
	assert.Equal(t, false, cfg.Blocking.Opa.Enabled.Value)
}
