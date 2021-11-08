package state

import (
	"log"
	"sync"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	"google.golang.org/protobuf/proto"
)

var (
	cfg    *traceableconfig.AgentConfig
	cfgMux = &sync.Mutex{}
)

// InitConfig initializes the config with default values
func InitConfig(c *traceableconfig.AgentConfig) {
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
	cfg, ok = proto.Clone(c).(*traceableconfig.AgentConfig)
	if !ok {
		log.Fatal("failed to initialize config.")
	}
}

// GetConfig returns the config value
func GetConfig() *traceableconfig.AgentConfig {
	if cfg == nil {
		InitConfig(config.Load().Blocking)
	}

	return cfg
}
