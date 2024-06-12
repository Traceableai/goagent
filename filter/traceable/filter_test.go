//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestLibTraceableAttributes(t *testing.T) {
	// nil
	libTraceableAttributes := createLibTraceableAttributes(nil)
	assert.Equal(t, 0, int(libTraceableAttributes.count))

	// empty map
	m := make(map[string]string)
	libTraceableAttributes = createLibTraceableAttributes(m)
	mFromLibTraceableAttributes := fromLibTraceableAttributes(libTraceableAttributes)
	assert.Equal(t, 0, int(libTraceableAttributes.count))
	assert.Equal(t, m, mFromLibTraceableAttributes)
	freeLibTraceableAttributes(libTraceableAttributes)

	// one element
	m["http.url"] = "https://www.foo.com/checkout?order_id=1"
	libTraceableAttributes = createLibTraceableAttributes(m)
	mFromLibTraceableAttributes = fromLibTraceableAttributes(libTraceableAttributes)
	assert.Equal(t, 1, int(libTraceableAttributes.count))
	assert.Equal(t, m, mFromLibTraceableAttributes)
	freeLibTraceableAttributes(libTraceableAttributes)

	// more than one element
	m["http.request.header.x-forwarded-for"] = "1.2.3.4"
	libTraceableAttributes = createLibTraceableAttributes(m)
	mFromLibTraceableAttributes = fromLibTraceableAttributes(libTraceableAttributes)
	assert.Equal(t, 2, int(libTraceableAttributes.count))
	assert.Equal(t, m, mFromLibTraceableAttributes)
	freeLibTraceableAttributes(libTraceableAttributes)
}

func TestTraceableConfigDisabled(t *testing.T) {
	f := NewFilter("test-service",
		&traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(false),
			},
			Sampling: &traceableconfig.SamplingConfig{
				Enabled: traceableconfig.Bool(false),
			},
		},
		zap.NewNop())
	assert.IsType(t, Filter{}, *f)
	assert.False(t, f.started)

	f.Start() // the blocking engine was not enabled thus start will never be true
	assert.False(t, f.started)
	f.Stop()
}

func TestGetLibTraceableConfig(t *testing.T) {
	libTraceableConfig := getLibTraceableConfig(
		"test-service",
		&traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(true),
				Modsecurity: &traceableconfig.ModsecurityConfig{
					Enabled: traceableconfig.Bool(false),
				},
				EvaluateBody:        traceableconfig.Bool(false),
				SkipInternalRequest: wrapperspb.Bool(false),
				RegionBlocking: &traceableconfig.RegionBlockingConfig{
					Enabled: traceableconfig.Bool(false),
				},
				MaxRecursionDepth: traceableconfig.Int32(10),
				RemoteConfig: &traceableconfig.RemoteConfig{
					Enabled:                traceableconfig.Bool(true),
					Endpoint:               traceableconfig.String("localhost:5441"),
					CertFile:               traceableconfig.String(""),
					PollPeriodSeconds:      traceableconfig.Int32(30),
					GrpcMaxCallRecvMsgSize: traceableconfig.Int32(32 * 1024 * 1024),
					UseSecureConnection:    traceableconfig.Bool(false),
				},
			},
			RemoteConfig: &traceableconfig.RemoteConfig{
				Enabled:                traceableconfig.Bool(true),
				Endpoint:               traceableconfig.String("localhost:5441"),
				CertFile:               traceableconfig.String("/conf/tls.crt"),
				PollPeriodSeconds:      traceableconfig.Int32(60),
				GrpcMaxCallRecvMsgSize: traceableconfig.Int32(64 * 1024 * 1024),
				UseSecureConnection:    traceableconfig.Bool(true),
			},
			Sampling: &traceableconfig.SamplingConfig{
				Enabled: traceableconfig.Bool(true),
				DefaultRateLimitConfig: &traceableconfig.RateLimitConfig{
					Enabled:               traceableconfig.Bool(false),
					MaxCountGlobal:        &wrapperspb.Int64Value{Value: 2},
					MaxCountPerEndpoint:   &wrapperspb.Int64Value{Value: 1},
					RefreshPeriod:         traceableconfig.String("30s"),
					ValueExpirationPeriod: traceableconfig.String("200h"),
					SpanType:              traceableconfig.SpanType_SPAN_TYPE_NO_SPAN,
				},
			},
			Logging: &traceableconfig.LogConfig{
				LogMode:  traceableconfig.LogMode_LOG_MODE_FILE,
				LogLevel: traceableconfig.LogLevel_LOG_LEVEL_WARN,
				LogFile: &traceableconfig.LogFileConfig{
					MaxFiles:    traceableconfig.Int32(10),
					MaxFileSize: traceableconfig.Int32(100 * 1024 * 1024),
					FilePath:    traceableconfig.String("/etc/libtraceable.log"),
				},
			},
			MetricsConfig: &traceableconfig.MetricsConfig{
				Enabled: traceableconfig.Bool(true),
				EndpointConfig: &traceableconfig.EndpointMetricsConfig{
					Enabled: traceableconfig.Bool(true),
					// same values from libtraceable defaults
					MaxEndpoints: traceableconfig.Int32(5000),
					Logging: &traceableconfig.MetricsLogConfig{
						Enabled:   traceableconfig.Bool(true),
						Frequency: traceableconfig.String("30m"),
					},
				},
				Logging: &traceableconfig.MetricsLogConfig{
					Enabled:   traceableconfig.Bool(true),
					Frequency: traceableconfig.String("30m"),
				},
			},
		},
	)

	assert.Equal(t, 1, int(libTraceableConfig.blocking_config.enabled))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.modsecurity_config.enabled))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.evaluate_body))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.skip_internal_request))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.rb_config.enabled))
	assert.Equal(t, 10, int(libTraceableConfig.blocking_config.max_recursion_depth))

	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "localhost:5441", getGoString(libTraceableConfig.remote_config.remote_endpoint))
	assert.Equal(t, "/conf/tls.crt", getGoString(libTraceableConfig.remote_config.cert_file))
	assert.Equal(t, int64(64*1024*1024), int64(libTraceableConfig.remote_config.grpc_max_call_recv_msg_size))
	assert.Equal(t, 60, int(libTraceableConfig.remote_config.poll_period_sec))
	assert.Equal(t, 1, int(libTraceableConfig.remote_config.use_secure_connection))

	assert.Equal(t, 1, int(libTraceableConfig.sampling_config.enabled))
	assert.Equal(t, 0, int(libTraceableConfig.sampling_config.default_rate_limit_config.enabled))
	assert.Equal(t, 2, int(libTraceableConfig.sampling_config.default_rate_limit_config.max_count_global))
	assert.Equal(t, 1, int(libTraceableConfig.sampling_config.default_rate_limit_config.max_count_per_endpoint))
	assert.Equal(t, "30s", getGoString(libTraceableConfig.sampling_config.default_rate_limit_config.refresh_period))
	assert.Equal(t, "200h",
		getGoString(libTraceableConfig.sampling_config.default_rate_limit_config.value_expiration_period))
	assert.Equal(t, traceableconfig.SpanType_SPAN_TYPE_NO_SPAN,
		getGoSpanType(libTraceableConfig.sampling_config.default_rate_limit_config.span_type))

	assert.Equal(t, traceableconfig.LogMode_LOG_MODE_FILE, getGoLogMode(libTraceableConfig.log_config.mode))
	assert.Equal(t, traceableconfig.LogLevel_LOG_LEVEL_WARN, getGoLogLevel(libTraceableConfig.log_config.level))
	assert.Equal(t, 10, int(libTraceableConfig.log_config.file_config.max_files))
	assert.Equal(t, 100*1024*1024, int(libTraceableConfig.log_config.file_config.max_file_size))
	assert.Equal(t, "/etc/libtraceable.log", getGoString(libTraceableConfig.log_config.file_config.log_file))

	assert.Equal(t, 1, int(libTraceableConfig.metrics_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.metrics_config.endpoint_config.enabled))
	assert.Equal(t, 5000, int(libTraceableConfig.metrics_config.endpoint_config.max_endpoints))
	assert.Equal(t, 1, int(libTraceableConfig.metrics_config.endpoint_config.logging.enabled))
	assert.Equal(t, "30m", getGoString(libTraceableConfig.metrics_config.endpoint_config.logging.frequency))
	assert.Equal(t, 1, int(libTraceableConfig.metrics_config.logging.enabled))
	assert.Equal(t, "30m", getGoString(libTraceableConfig.metrics_config.logging.frequency))

	assert.Equal(t, "", getGoString(libTraceableConfig.agent_config.environment))

	// verify for deprecated RemoteConfig
	libTraceableConfig = getLibTraceableConfig(
		"test-service",
		&traceableconfig.AgentConfig{
			DebugLog: traceableconfig.Bool(true), // ignored during parsing
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled:  traceableconfig.Bool(true),
				DebugLog: traceableconfig.Bool(true), //ignored during parsing
				Modsecurity: &traceableconfig.ModsecurityConfig{
					Enabled: traceableconfig.Bool(true),
				},
				EvaluateBody: traceableconfig.Bool(true),
				RegionBlocking: &traceableconfig.RegionBlockingConfig{
					Enabled: traceableconfig.Bool(true),
				},
				MaxRecursionDepth:   traceableconfig.Int32(10),
				SkipInternalRequest: traceableconfig.Bool(true),
				// takes precedence over top-level RemoteConfig
				RemoteConfig: &traceableconfig.RemoteConfig{
					Enabled:                traceableconfig.Bool(true),
					Endpoint:               traceableconfig.String("agent.traceableai:5441"),
					PollPeriodSeconds:      traceableconfig.Int32(10),
					CertFile:               traceableconfig.String("/etc/tls.crt"),
					GrpcMaxCallRecvMsgSize: traceableconfig.Int32(64 * 1024 * 1024),
					UseSecureConnection:    traceableconfig.Bool(true),
				},
			},
			// ignored during parsing
			Opa: &traceableconfig.Opa{
				Enabled:             traceableconfig.Bool(true),
				Endpoint:            traceableconfig.String("http://opa:8181"),
				PollPeriodSeconds:   traceableconfig.Int32(10),
				CertFile:            traceableconfig.String("/conf/tls.crt"),
				UseSecureConnection: traceableconfig.Bool(true),
			},
			RemoteConfig: &traceableconfig.RemoteConfig{
				Enabled:                traceableconfig.Bool(true),
				Endpoint:               traceableconfig.String("localhost:5441"),
				PollPeriodSeconds:      traceableconfig.Int32(30),
				CertFile:               traceableconfig.String(""),
				GrpcMaxCallRecvMsgSize: traceableconfig.Int32(32 * 1024 * 1024),
				UseSecureConnection:    traceableconfig.Bool(false),
			},
			Sampling: &traceableconfig.SamplingConfig{
				Enabled: traceableconfig.Bool(false),
				DefaultRateLimitConfig: &traceableconfig.RateLimitConfig{
					Enabled:               traceableconfig.Bool(false),
					MaxCountGlobal:        &wrapperspb.Int64Value{Value: 2},
					MaxCountPerEndpoint:   &wrapperspb.Int64Value{Value: 1},
					RefreshPeriod:         traceableconfig.String("30s"),
					ValueExpirationPeriod: traceableconfig.String("200h"),
					SpanType:              traceableconfig.SpanType_SPAN_TYPE_NO_SPAN,
				},
			},
			Logging: &traceableconfig.LogConfig{
				LogMode:  traceableconfig.LogMode_LOG_MODE_STDOUT,
				LogLevel: traceableconfig.LogLevel_LOG_LEVEL_INFO,
				LogFile: &traceableconfig.LogFileConfig{
					MaxFiles:    traceableconfig.Int32(3),
					MaxFileSize: traceableconfig.Int32(10 * 1024 * 1024),
					FilePath:    traceableconfig.String("/var/log/traceable/libtraceable-goagent.log"),
				},
			},
			MetricsConfig: &traceableconfig.MetricsConfig{
				Enabled: traceableconfig.Bool(true),
				EndpointConfig: &traceableconfig.EndpointMetricsConfig{
					Enabled: traceableconfig.Bool(true),
					// same values from libtraceable defaults
					MaxEndpoints: traceableconfig.Int32(5000),
					Logging: &traceableconfig.MetricsLogConfig{
						Enabled:   traceableconfig.Bool(true),
						Frequency: traceableconfig.String("30m"),
					},
				},
				Logging: &traceableconfig.MetricsLogConfig{
					Enabled:   traceableconfig.Bool(true),
					Frequency: traceableconfig.String("30m"),
				},
			},
		},
	)

	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "agent.traceableai:5441", getGoString(libTraceableConfig.remote_config.remote_endpoint))
	assert.Equal(t, 10, int(libTraceableConfig.remote_config.poll_period_sec))
	assert.Equal(t, "/etc/tls.crt", getGoString(libTraceableConfig.remote_config.cert_file))
	assert.Equal(t, int64(64*1024*1024), int64(libTraceableConfig.remote_config.grpc_max_call_recv_msg_size))
	assert.Equal(t, 1, int(libTraceableConfig.remote_config.use_secure_connection))

	assert.Equal(t, "", getGoString(libTraceableConfig.agent_config.environment))

	// Environment is present in agentConfig
	libTraceableConfig = getLibTraceableConfig(
		"test-service",
		&traceableconfig.AgentConfig{
			DebugLog: traceableconfig.Bool(true), // ignored during parsing
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled:  traceableconfig.Bool(true),
				DebugLog: traceableconfig.Bool(true), //ignored during parsing
				Modsecurity: &traceableconfig.ModsecurityConfig{
					Enabled: traceableconfig.Bool(true),
				},
				EvaluateBody: traceableconfig.Bool(true),
				RegionBlocking: &traceableconfig.RegionBlockingConfig{
					Enabled: traceableconfig.Bool(true),
				},
				MaxRecursionDepth:   traceableconfig.Int32(10),
				SkipInternalRequest: traceableconfig.Bool(true),
				// takes precedence over top-level RemoteConfig
				RemoteConfig: &traceableconfig.RemoteConfig{
					Enabled:                traceableconfig.Bool(true),
					Endpoint:               traceableconfig.String("agent.traceableai:5441"),
					PollPeriodSeconds:      traceableconfig.Int32(10),
					CertFile:               traceableconfig.String("/etc/tls.crt"),
					GrpcMaxCallRecvMsgSize: traceableconfig.Int32(64 * 1024 * 1024),
					UseSecureConnection:    traceableconfig.Bool(true),
				},
			},
			// ignored during parsing
			Opa: &traceableconfig.Opa{
				Enabled:             traceableconfig.Bool(true),
				Endpoint:            traceableconfig.String("http://opa:8181"),
				PollPeriodSeconds:   traceableconfig.Int32(10),
				CertFile:            traceableconfig.String("/conf/tls.crt"),
				UseSecureConnection: traceableconfig.Bool(true),
			},
			RemoteConfig: &traceableconfig.RemoteConfig{
				Enabled:                traceableconfig.Bool(true),
				Endpoint:               traceableconfig.String("localhost:5441"),
				PollPeriodSeconds:      traceableconfig.Int32(30),
				CertFile:               traceableconfig.String(""),
				GrpcMaxCallRecvMsgSize: traceableconfig.Int32(32 * 1024 * 1024),
				UseSecureConnection:    traceableconfig.Bool(false),
			},
			Sampling: &traceableconfig.SamplingConfig{
				Enabled: traceableconfig.Bool(false),
				DefaultRateLimitConfig: &traceableconfig.RateLimitConfig{
					Enabled:               traceableconfig.Bool(false),
					MaxCountGlobal:        &wrapperspb.Int64Value{Value: 2},
					MaxCountPerEndpoint:   &wrapperspb.Int64Value{Value: 1},
					RefreshPeriod:         traceableconfig.String("30s"),
					ValueExpirationPeriod: traceableconfig.String("200h"),
					SpanType:              traceableconfig.SpanType_SPAN_TYPE_NO_SPAN,
				},
			},
			Logging: &traceableconfig.LogConfig{
				LogMode:  traceableconfig.LogMode_LOG_MODE_STDOUT,
				LogLevel: traceableconfig.LogLevel_LOG_LEVEL_INFO,
				LogFile: &traceableconfig.LogFileConfig{
					MaxFiles:    traceableconfig.Int32(3),
					MaxFileSize: traceableconfig.Int32(10 * 1024 * 1024),
					FilePath:    traceableconfig.String("/var/log/traceable/libtraceable-goagent.log"),
				},
			},
			MetricsConfig: &traceableconfig.MetricsConfig{
				Enabled: traceableconfig.Bool(true),
				EndpointConfig: &traceableconfig.EndpointMetricsConfig{
					Enabled: traceableconfig.Bool(true),
					// same values from libtraceable defaults
					MaxEndpoints: traceableconfig.Int32(5000),
					Logging: &traceableconfig.MetricsLogConfig{
						Enabled:   traceableconfig.Bool(true),
						Frequency: traceableconfig.String("30m"),
					},
				},
				Logging: &traceableconfig.MetricsLogConfig{
					Enabled:   traceableconfig.Bool(true),
					Frequency: traceableconfig.String("30m"),
				},
			},
			Environment: traceableconfig.String("test-environment"),
		},
	)

	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "test-environment", getGoString(libTraceableConfig.agent_config.environment))
}
