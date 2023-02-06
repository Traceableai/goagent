package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadWithDefaults(t *testing.T) {
	cfg := Load()

	assert.Equal(t, "localhost:4317", cfg.Tracing.Reporting.Endpoint.Value)
	assert.Equal(t, false, cfg.Tracing.Reporting.Secure.Value)
	assert.Equal(t, TraceReporterType_OTLP, cfg.Tracing.Reporting.TraceReporterType)
	assert.Equal(t, true, cfg.TraceableConfig.Opa.Enabled.Value)
	assert.Equal(t, "http://localhost:8181/", cfg.TraceableConfig.Opa.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.Opa.PollPeriodSeconds.Value)

	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Request.Value)
	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Response.Value)
	assert.Equal(t, int32(131072), cfg.Tracing.DataCapture.BodyMaxSizeBytes.Value)
	assert.Equal(t, int32(1048576), cfg.Tracing.DataCapture.BodyMaxProcessingSizeBytes.Value)

	assert.Equal(t, false, cfg.TraceableConfig.DebugLog.Value)
	assert.Equal(t, true, cfg.TraceableConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "localhost:5441", cfg.TraceableConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "", cfg.TraceableConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)
	assert.Equal(t, false, cfg.TraceableConfig.DebugLog.Value)
	assert.Equal(t, true, cfg.TraceableConfig.ApiDiscovery.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.Sampling.Enabled.Value)

	// deprecated defaults
	assert.Equal(t, false, cfg.TraceableConfig.BlockingConfig.DebugLog.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "localhost:5441", cfg.TraceableConfig.BlockingConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.BlockingConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "", cfg.TraceableConfig.BlockingConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, "localhost:5441", cfg.TraceableConfig.BlockingConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.BlockingConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)
	assert.Equal(t, int32(403), cfg.TraceableConfig.BlockingConfig.ResponseStatusCode.Value)
}

func TestLoadFromFile(t *testing.T) {
	cfg := LoadFromFile("./testdata/config.yaml")

	assert.Equal(t, "goagent-example", cfg.Tracing.ServiceName.Value)
	assert.Equal(t, "traceable-agent:4317", cfg.Tracing.Reporting.Endpoint.Value)
	assert.Equal(t, false, cfg.Tracing.Reporting.Secure.Value)
	assert.Equal(t, TraceReporterType_OTLP, cfg.Tracing.Reporting.TraceReporterType)
	assert.Equal(t, true, cfg.TraceableConfig.Opa.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:8181/", cfg.TraceableConfig.Opa.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.Opa.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.TraceableConfig.Opa.CertFile.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.Enabled.Value)
	assert.Equal(t, false, cfg.TraceableConfig.BlockingConfig.Modsecurity.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.RegionBlocking.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.EvaluateBody.Value)
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.SkipInternalRequest.Value)
	assert.Equal(t, int32(472), cfg.TraceableConfig.BlockingConfig.ResponseStatusCode.Value)

	assert.Equal(t, true, cfg.TraceableConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:5441/", cfg.TraceableConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.TraceableConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)

	assert.Equal(t, true, cfg.TraceableConfig.ApiDiscovery.Enabled.Value)
	assert.Equal(t, true, cfg.TraceableConfig.Sampling.Enabled.Value)

	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Request.Value)
	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Response.Value)
}

func TestLoadFromFileDeprecatetd(t *testing.T) {
	cfg := LoadFromFile("./testdata/config_deprecated.yaml")

	// check debug_log in deprecated location
	assert.Equal(t, true, cfg.TraceableConfig.BlockingConfig.DebugLog.Value)

	// check remote_config in deprecated location still works
	assert.Equal(t, true, cfg.TraceableConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:5441/", cfg.TraceableConfig.BlockingConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.TraceableConfig.BlockingConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.TraceableConfig.BlockingConfig.RemoteConfig.CertFile.Value)
	assert.Equal(t, int32(32*1024*1024), cfg.TraceableConfig.BlockingConfig.RemoteConfig.GrpcMaxCallRecvMsgSize.Value)
}
