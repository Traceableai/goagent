package state

import (
	"log"
	"sync"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"google.golang.org/protobuf/proto"
)

var (
	cfg    *config.AgentConfig
	cfgMux = &sync.Mutex{}
)

// InitConfig initializes the config with default values
func InitConfig(c *config.AgentConfig) {
	cfgMux.Lock()
	defer cfgMux.Unlock()

	if cfg != nil {
		log.Println("config already initialized, ignoring new config.")
		return
	}

	// The reason why we clone the message instead of reusing the one passed by the user
	// is because user might decide to change values in runtime and that is undesirable
	// without a proper API.
	var ok bool
	tracingCfg, ok := proto.Clone(c.Tracing).(*hyperconfig.AgentConfig)
	if !ok {
		log.Fatal("failed to initialize hypertrace config.")
	}

	libtraceableCfg, ok := proto.Clone(c.TraceableConfig).(*traceableconfig.AgentConfig)
	if !ok {
		log.Fatal("failed to initialize traceable config.")
	}

	cfg = &config.AgentConfig{
		Tracing:         tracingCfg,
		TraceableConfig: libtraceableCfg,
	}
}

// GetConfig returns the config value
func GetConfig() *config.AgentConfig {
	if cfg == nil {
		InitConfig(config.Load())
	}

	return cfg
}
