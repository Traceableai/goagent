package config // import "github.com/Traceableai/goagent/config"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

const envPrefix = "TA_"

func Load() *AgentConfig {
	return &AgentConfig{
		Tracing: hyperconfig.Load(
			hyperconfig.WithEnvPrefix(envPrefix),
			hyperconfig.WithDefaults(defaultConfig.Tracing),
		),
		Blocking: traceableconfig.Load(
			traceableconfig.WithEnvPrefix(envPrefix),
			traceableconfig.WithDefaults(defaultConfig.Blocking),
		),
	}
}

func LoadFromFile(configFile string) *AgentConfig {
	return &AgentConfig{
		Tracing: hyperconfig.LoadFromFile(
			configFile,
			hyperconfig.WithEnvPrefix(envPrefix),
			hyperconfig.WithDefaults(defaultConfig.Tracing),
		),
		Blocking: traceableconfig.LoadFromFile(
			configFile,
			traceableconfig.WithEnvPrefix(envPrefix),
			traceableconfig.WithDefaults(defaultConfig.Blocking),
		),
	}
}
