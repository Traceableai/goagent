package config // import "github.com/Traceableai/goagent/config"

import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	hyperconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var defaultRemoteConfig = &traceableconfig.RemoteConfig{
	Enabled:                traceableconfig.Bool(true),
	Endpoint:               traceableconfig.String("localhost:5441"),
	PollPeriodSeconds:      traceableconfig.Int32(30),
	CertFile:               traceableconfig.String(""),
	GrpcMaxCallRecvMsgSize: traceableconfig.Int32(32 * 1024 * 1024),
}

// defaultConfig holds the default config values for agent.
var defaultConfig = &AgentConfig{
	Tracing: &hyperconfig.AgentConfig{
		Enabled:            hyperconfig.Bool(true),
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
			BodyMaxSizeBytes:           hyperconfig.Int32(131072),
			BodyMaxProcessingSizeBytes: hyperconfig.Int32(1048576),
			AllowedContentTypes: []*wrapperspb.StringValue{wrapperspb.String("json"),
				wrapperspb.String("x-www-form-urlencoded")},
		},
		Reporting: &hyperconfig.Reporting{
			Endpoint:          hyperconfig.String("localhost:4317"),
			Secure:            hyperconfig.Bool(false),
			TraceReporterType: hyperconfig.TraceReporterType_OTLP,
			CertFile:          hyperconfig.String(""),
		},
	},
	TraceableConfig: &traceableconfig.AgentConfig{
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
			RemoteConfig:        defaultRemoteConfig,
			ResponseStatusCode:  traceableconfig.Int32(403),
		},
		DebugLog:     traceableconfig.Bool(false),
		RemoteConfig: defaultRemoteConfig,
		ApiDiscovery: &traceableconfig.ApiDiscoveryConfig{
			Enabled: traceableconfig.Bool(true),
		},
		Sampling: &traceableconfig.SamplingConfig{
			Enabled: traceableconfig.Bool(true),
		},
	},
}

func GetDefaultRemoteConfig() *traceableconfig.RemoteConfig {
	return defaultRemoteConfig
}
