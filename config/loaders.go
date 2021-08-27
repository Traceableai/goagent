package config

import agentconfig "github.com/hypertrace/agent-config/gen/go/v1"

func Load() *AgentConfig {
	return &AgentConfig{
		Hypertrace: agentconfig.Load(
			agentconfig.WithEnvPrefix("TRACEABLE_"),
			agentconfig.WithDefaults(defaultConfig.Hypertrace),
		),
	}
}

func LoadFromFile(configFile string) *AgentConfig {
	return &AgentConfig{
		Hypertrace: agentconfig.LoadFromFile(
			configFile,
			agentconfig.WithEnvPrefix("TRACEABLE_"),
			agentconfig.WithDefaults(defaultConfig.Hypertrace),
		),
	}
}

func PropagationFormats(formats ...agentconfig.PropagationFormat) []agentconfig.PropagationFormat {
	return formats
}

var (
	Bool                           = agentconfig.Bool
	String                         = agentconfig.String
	Int32                          = agentconfig.Int32
	TraceReporterType_OTLP         = agentconfig.TraceReporterType_OTLP
	TraceReporterType_ZIPKIN       = agentconfig.TraceReporterType_ZIPKIN
	PropagationFormat_B3           = agentconfig.PropagationFormat_B3
	PropagationFormat_TRACECONTEXT = agentconfig.PropagationFormat_TRACECONTEXT
)
