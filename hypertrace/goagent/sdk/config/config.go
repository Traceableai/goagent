package config // import "github.com/Traceableai/goagent/hypertrace/goagent/sdk/config"

import (
	internalconfig "github.com/Traceableai/goagent/hypertrace/goagent/sdk/internal/config"
	agentconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

// InitConfig allows users to initialize the config
func InitConfig(c *agentconfig.AgentConfig) {
	internalconfig.InitConfig(c)
}

func ResetConfig() {
	internalconfig.ResetConfig()
}
