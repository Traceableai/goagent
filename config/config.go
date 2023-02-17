package config // import "github.com/Traceableai/goagent/config"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

type AgentConfig struct {
	Tracing         *hyperconfig.AgentConfig
	TraceableConfig *traceableconfig.AgentConfig
}

func LoadEnv(cfg *AgentConfig) {
	cfg.Tracing.LoadFromEnv(
		hyperconfig.WithEnvPrefix(envPrefix),
		hyperconfig.WithDefaults(defaultConfig.Tracing),
	)

	cfg.TraceableConfig.LoadFromEnv(
		traceableconfig.WithEnvPrefix(envPrefix),
		traceableconfig.WithDefaults(defaultConfig.TraceableConfig),
	)
}

func PropagationFormats(formats ...hyperconfig.PropagationFormat) []hyperconfig.PropagationFormat {
	return formats
}

var (
	Bool                                            = traceableconfig.Bool
	String                                          = traceableconfig.String
	Int32                                           = traceableconfig.Int32
	TraceReporterType_OTLP                          = hyperconfig.TraceReporterType_OTLP
	TraceReporterType_ZIPKIN                        = hyperconfig.TraceReporterType_ZIPKIN
	PropagationFormat_B3                            = hyperconfig.PropagationFormat_B3
	PropagationFormat_TRACECONTEXT                  = hyperconfig.PropagationFormat_TRACECONTEXT
	MetricReporterType_METRIC_REPORTER_TYPE_LOGGING = hyperconfig.MetricReporterType_METRIC_REPORTER_TYPE_LOGGING
	MetricReporterType_METRIC_REPORTER_TYPE_OTLP    = hyperconfig.MetricReporterType_METRIC_REPORTER_TYPE_OTLP
)
