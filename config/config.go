package config // import "github.com/Traceableai/goagent/config"

import (
	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

type AgentConfig struct {
	Tracing  *hyperconfig.AgentConfig
	Blocking *traceconfig.AgentConfig
}
