package state

import (
	"testing"

	config "github.com/Traceableai/agent-config/gen/go/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConfig(t *testing.T) {
	InitConfig(&config.AgentConfig{
		Opa: &config.Opa{
			Enabled: wrapperspb.Bool(true),
		},
	})
}
