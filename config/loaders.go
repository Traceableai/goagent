package config // import "github.com/Traceableai/goagent/config"

import (
	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

const envPrefix = "TRACEABLE_"

func Load() *AgentConfig {
	return &AgentConfig{
		Tracing: hyperconfig.Load(
			hyperconfig.WithEnvPrefix(envPrefix),
			hyperconfig.WithDefaults(defaultConfig.Tracing),
		),
		Blocking: traceconfig.Load(
			traceconfig.WithEnvPrefix(envPrefix),
			traceconfig.WithDefaults(defaultConfig.Blocking),
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
		Blocking: traceconfig.LoadFromFile(
			configFile,
			traceconfig.WithEnvPrefix(envPrefix),
			traceconfig.WithDefaults(defaultConfig.Blocking),
		),
	}
}
