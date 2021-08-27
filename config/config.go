package config

import (
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
	traceconfig "github.com/traceableai/agent-config/gen/go/v1"
)

type AgentConfig struct {
	Hypertrace *hyperconfig.AgentConfig
	Traceable  traceconfig.Traceable
}
