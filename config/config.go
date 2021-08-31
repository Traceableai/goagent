package config // import "github.com/Traceableai/goagent/config"

import (
	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

type AgentConfig struct {
	Tracing  *hyperconfig.AgentConfig
	Blocking *traceconfig.AgentConfig
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
