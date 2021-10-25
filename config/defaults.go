package config // import "github.com/Traceableai/goagent/config"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
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
			Endpoint:          hyperconfig.String("localhost:4317"),
			Secure:            hyperconfig.Bool(false),
			TraceReporterType: hyperconfig.TraceReporterType_OTLP,
			CertFile:          hyperconfig.String(""),
		},
	},
	Blocking: &traceableconfig.AgentConfig{
		Opa: &traceableconfig.Opa{
			Enabled:           traceableconfig.Bool(true),
			Endpoint:          traceableconfig.String("http://localhost:8181/"),
			PollPeriodSeconds: traceableconfig.Int32(30),
			CertFile:          traceableconfig.String(""),
		},
		BlockingConfig: &traceableconfig.BlockingConfig{
			Enabled:  traceableconfig.Bool(true),
			DebugLog: traceableconfig.Bool(false),
			Modsecurity: &traceableconfig.ModsecurityConfig{
				Enabled: traceableconfig.Bool(true),
			},
			EvaluateBody: traceableconfig.Bool(true),
			RegionBlocking: &traceableconfig.RegionBlockingConfig{
				Enabled: traceableconfig.Bool(true),
			},
			SkipInternalRequest: traceableconfig.Bool(true),
			RemoteConfig: &traceableconfig.RemoteConfig{
				Enabled:           traceableconfig.Bool(true),
				Endpoint:          traceableconfig.String("localhost:5441"),
				PollPeriodSeconds: traceableconfig.Int32(30),
				CertFile:          traceableconfig.String(""),
			},
		},
	},
}
