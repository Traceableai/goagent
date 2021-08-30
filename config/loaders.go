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

func PropagationFormats(formats ...hyperconfig.PropagationFormat) []hyperconfig.PropagationFormat {
	return formats
}

var (
	Bool                           = traceconfig.Bool
	String                         = traceconfig.String
	Int32                          = traceconfig.Int32
	TraceReporterType_OTLP         = hyperconfig.TraceReporterType_OTLP
	TraceReporterType_ZIPKIN       = hyperconfig.TraceReporterType_ZIPKIN
	PropagationFormat_B3           = hyperconfig.PropagationFormat_B3
	PropagationFormat_TRACECONTEXT = hyperconfig.PropagationFormat_TRACECONTEXT
)
