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
	assert.Equal(t, true, cfg.Blocking.Opa.Enabled.Value)
	assert.Equal(t, "http://localhost:8181/", cfg.Blocking.Opa.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.Blocking.Opa.PollPeriodSeconds.Value)

	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Request.Value)
	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Response.Value)
	assert.Equal(t, int32(131072), cfg.Tracing.DataCapture.BodyMaxSizeBytes.Value)
	assert.Equal(t, int32(1048576), cfg.Tracing.DataCapture.BodyMaxProcessingSizeBytes.Value)
}

func TestLoadFromFile(t *testing.T) {
	cfg := LoadFromFile("./testdata/config.yaml")

	assert.Equal(t, "goagent-example", cfg.Tracing.ServiceName.Value)
	assert.Equal(t, "traceable-agent:4317", cfg.Tracing.Reporting.Endpoint.Value)
	assert.Equal(t, false, cfg.Tracing.Reporting.Secure.Value)
	assert.Equal(t, TraceReporterType_OTLP, cfg.Tracing.Reporting.TraceReporterType)
	assert.Equal(t, true, cfg.Blocking.Opa.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:8181/", cfg.Blocking.Opa.Endpoint.Value)
	assert.Equal(t, int32(30), cfg.Blocking.Opa.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.Blocking.Opa.CertFile.Value)
	assert.Equal(t, true, cfg.Blocking.BlockingConfig.Enabled.Value)
	assert.Equal(t, true, cfg.Blocking.BlockingConfig.RemoteConfig.Enabled.Value)
	assert.Equal(t, "http://traceable-agent:5441/", cfg.Blocking.BlockingConfig.RemoteConfig.Endpoint.Value)
	assert.Equal(t, int32(60), cfg.Blocking.BlockingConfig.RemoteConfig.PollPeriodSeconds.Value)
	assert.Equal(t, "/conf/tls.crt", cfg.Blocking.BlockingConfig.RemoteConfig.CertFile.Value)

	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Request.Value)
	assert.Equal(t, true, cfg.Tracing.DataCapture.HttpBody.Response.Value)
}
