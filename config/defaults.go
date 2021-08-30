package config // import "github.com/Traceableai/goagent/config"

import (
	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
)

// defaultConfig holds the default config values for agent.
var defaultConfig = &AgentConfig{
	Tracing: &hyperconfig.AgentConfig{
		PropagationFormats: []hyperconfig.PropagationFormat{hyperconfig.PropagationFormat_TRACECONTEXT},
		DataCapture: &hyperconfig.DataCapture{
			HttpHeaders: &hyperconfig.Message{
				Request:  hyperconfig.Bool(true),
				Response: hyperconfig.Bool(true),
			},
			HttpBody: &hyperconfig.Message{
				Request:  hyperconfig.Bool(true),
				Response: hyperconfig.Bool(true),
			},
			RpcMetadata: &hyperconfig.Message{
				Request:  hyperconfig.Bool(true),
				Response: hyperconfig.Bool(true),
			},
			RpcBody: &hyperconfig.Message{
				Request:  hyperconfig.Bool(true),
				Response: hyperconfig.Bool(true),
			},
			BodyMaxSizeBytes: hyperconfig.Int32(131072),
		},
		Reporting: &hyperconfig.Reporting{
			Endpoint:          hyperconfig.String("http://localhost:9411/api/v2/spans"),
			Secure:            hyperconfig.Bool(false),
			TraceReporterType: hyperconfig.TraceReporterType_OTLP,
		},
	},
	Blocking: &traceconfig.AgentConfig{
		Opa: &traceconfig.Opa{
			Enabled:           traceconfig.Bool(true),
			Endpoint:          traceconfig.String("http://localhost:8181/"),
			PollPeriodSeconds: traceconfig.Int32(30),
		},
		BlockingConfig: &traceconfig.BlockingConfig{
			Enabled:  traceconfig.Bool(true),
			DebugLog: traceconfig.Bool(false),
			Modsecurity: &traceconfig.ModsecurityConfig{
				Enabled: traceconfig.Bool(true),
			},
			EvaluateBody: traceconfig.Bool(true),
			RegionBlocking: &traceconfig.RegionBlockingConfig{
				Enabled: traceconfig.Bool(true),
			},
			RemoteConfig: &traceconfig.RemoteConfig{
				Enabled:           traceconfig.Bool(true),
				Endpoint:          traceconfig.String("localhost:5441"),
				PollPeriodSeconds: traceconfig.Int32(30),
			},
		},
	},
}
