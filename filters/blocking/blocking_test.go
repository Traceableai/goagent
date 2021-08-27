//+build linux

package blocking

// IMPORTANT: These tests require the librtraceable.so library to be loaded
// in the host for which we run `make install-libs`, otherwise we need to compile
// the tests and run them linking the dynamic library as in Dockerfile.modsec.test

import (
	"testing"

	traceconfig "github.com/traceableai/agent-config/gen/go/v1"

	"github.com/hypertrace/goagent/sdk/filter"
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
	f := NewBlockingFilter(&traceconfig.AgentConfig{})
	assert.IsType(t, filter.NoopFilter{}, f)

	f = NewBlockingFilter(&traceconfig.AgentConfig
		BlockingConfig: &commonv1.BlockingConfig{
			Enabled: wrapperspb.Bool(false),
		},
	})
	assert.IsType(t, filter.NoopFilter{}, f)
}

func TestGetLibTraceableConfig(t *testing.T) {
	// most straightforward config, only opa is specified
	libTraceableConfig := getLibTraceableConfig(
		&traceconfig.AgentConfig{
			BlockingConfig{
				Enabled: wrapperspb.Bool(true),
			},
			Opa{
				Endpoint:          wrapperspb.String("http://opa:8181"),
				PollPeriodSeconds: wrapperspb.Int32(10),
			}
		}
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
		&traceconfig.AgentConfig {
		BlockingConfig: {
			Enabled:  wrapperspb.Bool(true),
			DebugLog: wrapperspb.Bool(true),
			Modsecurity: &commonv1.ModsecurityConfig{
				Enabled: wrapperspb.Bool(false),
			},
			EvaluateBody: wrapperspb.Bool(false),
			RegionBlocking: &commonv1.RegionBlockingConfig{
				Enabled: wrapperspb.Bool(false),
			},
			RemoteConfig: &commonv1.RemoteConfig{
				Enabled: wrapperspb.Bool(false),
			},
		},
		Opa: {
			Endpoint:          wrapperspb.String("http://opa:8181"),
			PollPeriodSeconds: wrapperspb.Int32(10),
		},
	}
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
		&traceconfig.AgentConfig {
		BlockingConfig: {
			Enabled: wrapperspb.Bool(true),
			RemoteConfig: &commonv1.RemoteConfig{
				Endpoint:          wrapperspb.String("agent.traceableai:5441"),
				PollPeriodSeconds: wrapperspb.Int32(10),
			},
		},
		Opa: {
			Endpoint:          wrapperspb.String("http://opa:8181"),
			PollPeriodSeconds: wrapperspb.Int32(10),
		},
	}
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
