package config

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
)

var DisabledConfig = &traceableconfig.AgentConfig{
	Opa: &traceableconfig.Opa{
		Enabled: config.Bool(false),
	},
	BlockingConfig: &traceableconfig.BlockingConfig{
		Enabled: config.Bool(false),
		Modsecurity: &traceableconfig.ModsecurityConfig{
			Enabled: config.Bool(false),
		},
		RemoteConfig: &traceableconfig.RemoteConfig{
			Enabled: config.Bool(false),
		},
		RegionBlocking: &traceableconfig.RegionBlockingConfig{
			Enabled: config.Bool(false),
		},
	},
}
