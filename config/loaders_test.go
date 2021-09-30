package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	cfg := LoadFromFile("./testdata/config.yaml")

	assert.Equal(t, cfg.Tracing.ServiceName.Value, "goagent-example")
	assert.Equal(t, cfg.Tracing.Reporting.Endpoint.Value, "traceable-agent:4317")
	assert.Equal(t, cfg.Tracing.Reporting.Secure.Value, false)
	assert.Equal(t, cfg.Tracing.Reporting.TraceReporterType, TraceReporterType_OTLP)
	assert.Equal(t, cfg.Blocking.Opa.Enabled.Value, true)
	assert.Equal(t, cfg.Blocking.Opa.Endpoint.Value, "http://traceable-agent:8181/")
	assert.Equal(t, cfg.Blocking.Opa.PollPeriodSeconds.Value, int32(30))

	assert.Equal(t, cfg.Tracing.DataCapture.HttpBody.Request.Value, true)
	assert.Equal(t, cfg.Tracing.DataCapture.HttpBody.Response.Value, true)
}
