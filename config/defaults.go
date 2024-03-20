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
	UseSecureConnection:    traceableconfig.Bool(false),
}

// defaultConfig holds the default config values for agent.
var defaultConfig = &AgentConfig{
	// TODO update ht agent config so that we can refer that directly
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
			AllowedContentTypes: []*wrapperspb.StringValue{
				wrapperspb.String("json"),
				wrapperspb.String("x-www-form-urlencoded"),
				wrapperspb.String("xml"),
			},
		},
		Reporting: &hyperconfig.Reporting{
			Endpoint:                hyperconfig.String("localhost:4317"),
			Secure:                  hyperconfig.Bool(false),
			TraceReporterType:       hyperconfig.TraceReporterType_OTLP,
			CertFile:                hyperconfig.String(""),
			EnableGrpcLoadbalancing: hyperconfig.Bool(true),
		},
	},
	TraceableConfig: &traceableconfig.AgentConfig{
		BlockingConfig: &traceableconfig.BlockingConfig{
			Enabled: traceableconfig.Bool(true),
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
			ResponseMessage:     traceableconfig.String("Access Forbidden"),
			MaxRecursionDepth:   traceableconfig.Int32(20),
		},
		RemoteConfig: defaultRemoteConfig,
		ApiDiscovery: &traceableconfig.ApiDiscoveryConfig{
			Enabled: traceableconfig.Bool(true),
		},
		Sampling: &traceableconfig.SamplingConfig{
			Enabled: traceableconfig.Bool(true),
			DefaultRateLimitConfig: &traceableconfig.RateLimitConfig{
				Enabled: traceableconfig.Bool(false),
			},
		},
		Logging: &traceableconfig.LogConfig{
			LogMode:  traceableconfig.LogMode_LOG_MODE_STDOUT,
			LogLevel: traceableconfig.LogLevel_LOG_LEVEL_INFO,
			LogFile: &traceableconfig.LogFileConfig{
				MaxFiles:    traceableconfig.Int32(3),
				MaxFileSize: traceableconfig.Int32(10485760),
				FilePath:    traceableconfig.String("/var/traceable/log/libtraceable-goagent.log"),
			},
		},
	},
}

func GetDefaultRemoteConfig() *traceableconfig.RemoteConfig {
	return defaultRemoteConfig
}
