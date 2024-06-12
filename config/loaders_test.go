package config

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/stretchr/testify/assert"
)

func TestLoadWithDefaults(t *testing.T) {
	cfg := Load()

	assert.Equal(t, "localhost:4317", cfg.Tracing.Reporting.Endpoint.Value)
	assert.Equal(t, false, cfg.Tracing.Reporting.Secure.Value)
	assert.Equal(t, TraceReporterType_OTLP, cfg.Tracing.Reporting.TraceReporterType)
	assert.Equal(t, true, cfg.Tracing.Reporting.EnableGrpcLoadbalancing.Value)

	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Request.Value)
	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Response.Value)
	assert.Equal(t, int32(131072), cfg.Tracing.DataCapture.BodyMaxSizeBytes.Value)
	assert.Equal(t, int32(1048576), cfg.Tracing.DataCapture.BodyMaxProcessingSizeBytes.Value)

	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.Modsecurity.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.EvaluateBody.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.RegionBlocking.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.SkipInternalRequest.Value)
	assert.Equal(t, int32(403), cfg.TraceableConfig.BlockingConfig.ResponseStatusCode.Value)
	assert.Equal(t, "Access Forbidden", cfg.TraceableConfig.BlockingConfig.ResponseMessage.Value)
	assert.Equal(t, int32(20), cfg.TraceableConfig.BlockingConfig.MaxRecursionDepth.Value)

	assert.Equal(t, true, cfg.TraceableConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "localhost:5441", cfg.TraceableConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "", cfg.TraceableConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)
	assert.Equal(t, false, cfg.TraceableConfig.RemoteConfig.UseSecureConnection.Value)

	assert.Equal(t, true, cfg.TraceableConfig.Sampling.Enabled.Value)
	assert.Equal(t, false, cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.Enabled.Value)

	assert.Equal(t, traceableconfig.LogMode_LOG_MODE_STDOUT, cfg.TraceableConfig.Logging.LogMode)
	assert.Equal(t, traceableconfig.LogLevel_LOG_LEVEL_INFO, cfg.TraceableConfig.Logging.LogLevel)
	assert.Equal(t, int32(3), cfg.TraceableConfig.Logging.LogFile.MaxFiles.Value)
	assert.Equal(t, int32(10485760), cfg.TraceableConfig.Logging.LogFile.MaxFileSize.Value)
	assert.Equal(t, "/var/traceable/log/libtraceable-goagent.log", cfg.TraceableConfig.Logging.LogFile.FilePath.Value)

	assert.False(t, cfg.TraceableConfig.MetricsConfig.Enabled.Value)
	assert.False(t, cfg.TraceableConfig.MetricsConfig.EndpointConfig.Enabled.Value)
	assert.Equal(t, int32(5000), cfg.TraceableConfig.MetricsConfig.EndpointConfig.MaxEndpoints.Value)
	assert.True(t, cfg.TraceableConfig.MetricsConfig.EndpointConfig.Logging.Enabled.Value)
	assert.Equal(t, "30m", cfg.TraceableConfig.MetricsConfig.EndpointConfig.Logging.Frequency.Value)
	assert.True(t, cfg.TraceableConfig.MetricsConfig.Logging.Enabled.Value)
	assert.Equal(t, "30m", cfg.TraceableConfig.MetricsConfig.Logging.Frequency.Value)

	// deprecated defaults
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "localhost:5441", cfg.TraceableConfig.BlockingConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.BlockingConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "", cfg.TraceableConfig.BlockingConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.BlockingConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)
	assert.Equal(t, false, cfg.TraceableConfig.BlockingConfig.RemoteConfig.UseSecureConnection.Value)

	// environment field check, has to be default
	assert.Equal(t, "", cfg.TraceableConfig.Environment.Value)
}

func TestLoadFromFile(t *testing.T) {
	cfg := LoadFromFile("./testdata/config.yaml")

	assert.Equal(t, "goagent-example", cfg.Tracing.ServiceName.Value)
	assert.Equal(t, "traceable-agent:4317", cfg.Tracing.Reporting.Endpoint.Value)
	assert.Equal(t, false, cfg.Tracing.Reporting.Secure.Value)
	assert.Equal(t, TraceReporterType_OTLP, cfg.Tracing.Reporting.TraceReporterType)

	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Request.Value)
	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Response.Value)

	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.Enabled.Value)
	assert.Equal(t, false, cfg.TraceableConfig.BlockingConfig.Modsecurity.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.RegionBlocking.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.EvaluateBody.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.SkipInternalRequest.Value)
	assert.Equal(t, int32(472), cfg.TraceableConfig.BlockingConfig.ResponseStatusCode.Value)
	assert.Equal(t, "Custom Forbidden Message", cfg.TraceableConfig.BlockingConfig.ResponseMessage.Value)

	assert.Equal(t, true, cfg.TraceableConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:5441/", cfg.TraceableConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.TraceableConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)

	assert.Equal(t, true, cfg.TraceableConfig.Sampling.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.Enabled.Value)
	assert.Equal(t, int64(9223372036854775807), cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.MaxCountGlobal.Value)
	assert.Equal(t, int64(3), cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.MaxCountPerEndpoint.Value)
	assert.Equal(t, "30s", cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.RefreshPeriod.Value)
	assert.Equal(t, "168h", cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.ValueExpirationPeriod.Value)
	assert.Equal(t, traceableconfig.SpanType_SPAN_TYPE_NO_SPAN, cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.SpanType)

	assert.Equal(t, traceableconfig.LogMode_LOG_MODE_FILE, cfg.TraceableConfig.Logging.LogMode)
	assert.Equal(t, traceableconfig.LogLevel_LOG_LEVEL_DEBUG, cfg.TraceableConfig.Logging.LogLevel)
	assert.Equal(t, int32(1), cfg.TraceableConfig.Logging.LogFile.MaxFiles.Value)
	assert.Equal(t, int32(104857600), cfg.TraceableConfig.Logging.LogFile.MaxFileSize.Value)
	assert.Equal(t, "/var/traceable/log/libtraceable-goagent.log", cfg.TraceableConfig.Logging.LogFile.FilePath.Value)

	// environment field check
	assert.Equal(t, "goagent-env", cfg.TraceableConfig.Environment.Value)
}

// To check if environment field is not provided in config file it should get default value
func TestLoadFromFileWithoutEnvironmentInConfig(t *testing.T) {
	cfg := LoadFromFile("./testdata/config-no-env.yaml")

	assert.Equal(t, "goagent-example", cfg.Tracing.ServiceName.Value)

	assert.Equal(t, true, cfg.TraceableConfig.Sampling.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.Enabled.Value)
	assert.Equal(t, int64(9223372036854775807), cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.MaxCountGlobal.Value)
	assert.Equal(t, int64(3), cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.MaxCountPerEndpoint.Value)
	assert.Equal(t, "30s", cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.RefreshPeriod.Value)
	assert.Equal(t, "168h", cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.ValueExpirationPeriod.Value)
	assert.Equal(t, traceableconfig.SpanType_SPAN_TYPE_NO_SPAN, cfg.TraceableConfig.Sampling.DefaultRateLimitConfig.SpanType)

	assert.Equal(t, traceableconfig.LogMode_LOG_MODE_FILE, cfg.TraceableConfig.Logging.LogMode)
	assert.Equal(t, traceableconfig.LogLevel_LOG_LEVEL_DEBUG, cfg.TraceableConfig.Logging.LogLevel)
	assert.Equal(t, int32(1), cfg.TraceableConfig.Logging.LogFile.MaxFiles.Value)
	assert.Equal(t, int32(104857600), cfg.TraceableConfig.Logging.LogFile.MaxFileSize.Value)
	assert.Equal(t, "/var/traceable/log/libtraceable-goagent.log", cfg.TraceableConfig.Logging.LogFile.FilePath.Value)

	// environment field check to be default
	assert.Equal(t, "", cfg.TraceableConfig.Environment.Value)
}

func TestLoadFromFileDeprecated(t *testing.T) {
	cfg := LoadFromFile("./testdata/config_deprecated.yaml")

	// check remote_config in deprecated location still works
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:5441/", cfg.TraceableConfig.BlockingConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.BlockingConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.TraceableConfig.BlockingConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.BlockingConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)
	assert.Equal(t, false, cfg.TraceableConfig.BlockingConfig.RemoteConfig.UseSecureConnection.Value)
}
