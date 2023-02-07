package config

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

var DisabledConfig = &config.AgentConfig{
	Tracing: &hyperconfig.AgentConfig{
		Enabled: config.Bool(false),
	},
	TraceableConfig: &traceableconfig.AgentConfig{
		Opa: &traceableconfig.Opa{
			Enabled: config.Bool(false),
		},
		BlockingConfig: &traceableconfig.BlockingConfig{
			Enabled: config.Bool(false),
			Modsecurity: &traceableconfig.ModsecurityConfig{
				Enabled: config.Bool(false),
			},

			RegionBlocking: &traceableconfig.RegionBlockingConfig{
				Enabled: config.Bool(false),
			},
		},
		RemoteConfig: &traceableconfig.RemoteConfig{
			Enabled: config.Bool(false),
		},
		ApiDiscovery: &traceableconfig.ApiDiscoveryConfig{
			Enabled: config.Bool(false),
		},
		Sampling: &traceableconfig.SamplingConfig{
			Enabled: config.Bool(false),
		},
	},
}
