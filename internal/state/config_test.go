package state

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/config"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConfig(t *testing.T) {
	InitConfig(&config.AgentConfig{
		TraceableConfig: &traceableconfig.AgentConfig{
			DetectionConfig: &traceableconfig.ThreatActivityDetection{
				Enabled: wrapperspb.Bool(true),
			},
		},
	})
}
