//go:build linux
// +build linux

package traceable

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
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
	m["http.url"] = "http://www.foo.com/checkout?order_id=1"
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

func TestBlockingDisabled(t *testing.T) {
	f := NewFilter(&traceableconfig.AgentConfig{}, zap.NewNop())
	assert.IsType(t, Filter{}, *f)
	assert.False(t, f.started)

	f.Start() // the blocking engine was not enabled thus start will never be true
	assert.False(t, f.started)
	f.Stop()

	f = NewFilter(&traceableconfig.AgentConfig{
		BlockingConfig: &traceableconfig.BlockingConfig{
			Enabled: wrapperspb.Bool(false),
		},
	}, zap.NewNop())
	assert.IsType(t, Filter{}, *f)
	assert.False(t, f.started)

	f.Start() // the blocking engine was not enabled thus start will never be true
	assert.False(t, f.started)
	f.Stop()
}

func TestGetLibTraceableConfig(t *testing.T) {
	// most straightforward config, only opa is specified
	libTraceableConfig := getLibTraceableConfig(
		&traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(true),
			},
			Opa: &traceableconfig.Opa{
				Endpoint:          traceableconfig.String("http://opa:8181"),
				PollPeriodSeconds: traceableconfig.Int32(10),
			},
		},
	)

	assert.Equal(t, "http://opa:8181", getGoString(libTraceableConfig.opa_config.opa_server_url))
	assert.Equal(t, 0, int(libTraceableConfig.log_config.mode))
	assert.Equal(t, 10, int(libTraceableConfig.opa_config.min_delay))
	assert.Equal(t, 10, int(libTraceableConfig.opa_config.max_delay))
	assert.Equal(t, 1, int(libTraceableConfig.opa_config.log_to_console))
	assert.Equal(t, 0, int(libTraceableConfig.opa_config.debug_log))
	assert.Equal(t, 0, int(libTraceableConfig.opa_config.skip_verify))
	assert.Equal(t, 1, int(libTraceableConfig.modsecurity_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.evaluate_body))
	assert.Equal(t, 1, int(libTraceableConfig.rb_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "localhost:5441", getGoString(libTraceableConfig.remote_config.remote_endpoint))
	assert.Equal(t, 30, int(libTraceableConfig.remote_config.poll_period_sec))

	// specify all options
	libTraceableConfig = getLibTraceableConfig(
		&traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled:  traceableconfig.Bool(true),
				DebugLog: traceableconfig.Bool(true),
				Modsecurity: &traceableconfig.ModsecurityConfig{
					Enabled: traceableconfig.Bool(false),
				},
				EvaluateBody: traceableconfig.Bool(false),
				RegionBlocking: &traceableconfig.RegionBlockingConfig{
					Enabled: traceableconfig.Bool(false),
				},
				RemoteConfig: &traceableconfig.RemoteConfig{
					Enabled: traceableconfig.Bool(false),
				},
			},
			Opa: &traceableconfig.Opa{
				Endpoint:          wrapperspb.String("http://opa:8181"),
				PollPeriodSeconds: wrapperspb.Int32(10),
			},
		},
	)

	assert.Equal(t, "http://opa:8181", getGoString(libTraceableConfig.opa_config.opa_server_url))
	assert.Equal(t, 1, int(libTraceableConfig.log_config.mode))
	assert.Equal(t, 10, int(libTraceableConfig.opa_config.min_delay))
	assert.Equal(t, 10, int(libTraceableConfig.opa_config.max_delay))
	assert.Equal(t, 1, int(libTraceableConfig.opa_config.log_to_console))
	assert.Equal(t, 1, int(libTraceableConfig.opa_config.debug_log))
	assert.Equal(t, 0, int(libTraceableConfig.opa_config.skip_verify))
	assert.Equal(t, 0, int(libTraceableConfig.modsecurity_config.enabled))
	assert.Equal(t, 0, int(libTraceableConfig.evaluate_body))
	assert.Equal(t, 0, int(libTraceableConfig.rb_config.enabled))
	assert.Equal(t, 0, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "", getGoString(libTraceableConfig.remote_config.remote_endpoint))

	// verify for RemoteConfig
	libTraceableConfig = getLibTraceableConfig(
		&traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(true),
				RemoteConfig: &traceableconfig.RemoteConfig{
					Endpoint:          traceableconfig.String("agent.traceableai:5441"),
					PollPeriodSeconds: traceableconfig.Int32(10),
				},
			},
			Opa: &traceableconfig.Opa{
				Endpoint:          traceableconfig.String("http://opa:8181"),
				PollPeriodSeconds: traceableconfig.Int32(10),
			},
		},
	)

	assert.Equal(t, "http://opa:8181", getGoString(libTraceableConfig.opa_config.opa_server_url))
	assert.Equal(t, 0, int(libTraceableConfig.log_config.mode))
	assert.Equal(t, 10, int(libTraceableConfig.opa_config.min_delay))
	assert.Equal(t, 10, int(libTraceableConfig.opa_config.max_delay))
	assert.Equal(t, 1, int(libTraceableConfig.opa_config.log_to_console))
	assert.Equal(t, 0, int(libTraceableConfig.opa_config.debug_log))
	assert.Equal(t, 0, int(libTraceableConfig.opa_config.skip_verify))
	assert.Equal(t, 1, int(libTraceableConfig.modsecurity_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.evaluate_body))
	assert.Equal(t, 1, int(libTraceableConfig.rb_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "agent.traceableai:5441", getGoString(libTraceableConfig.remote_config.remote_endpoint))
	assert.Equal(t, 10, int(libTraceableConfig.remote_config.poll_period_sec))
}
