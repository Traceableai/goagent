package config

import (
	"log"
	"sync"

	"github.com/hypertrace/goagent/config"
	traceconfig "github.com/traceableai/agent-config/gen/go/v1"
	"google.golang.org/protobuf/proto"
)

var cfg *traceconfig.Traceable
var cfgMux = &sync.Mutex{}

// InitConfig initializes the config with default values
func InitConfig(c *traceconfig.AgentConfig) {
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
	cfg, ok = proto.Clone(c).(*traceconfig.AgentConfig)
	if !ok {
		log.Fatal("failed to initialize config.")
	}
}

// GetConfig returns the config value
func GetConfig() *traceconfig.AgentConfig {
	if cfg == nil {
		InitConfig(config.Load())
	}

	return cfg
}
