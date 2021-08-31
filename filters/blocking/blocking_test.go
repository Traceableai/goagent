//+build linux

package blocking

import (
	"testing"

	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"

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

	f = NewBlockingFilter(&traceconfig.AgentConfig{
		BlockingConfig: &traceconfig.BlockingConfig{
			Enabled: wrapperspb.Bool(false),
		},
	})
	assert.IsType(t, filter.NoopFilter{}, f)
}

func TestGetLibTraceableConfig(t *testing.T) {
	// most straightforward config, only opa is specified
	libTraceableConfig := getLibTraceableConfig(
		&traceconfig.AgentConfig{
			BlockingConfig: &traceconfig.BlockingConfig{
				Enabled: traceconfig.Bool(true),
			},
			Opa: &traceconfig.Opa{
				Endpoint:          traceconfig.String("http://opa:8181"),
				PollPeriodSeconds: traceconfig.Int32(10),
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
		&traceconfig.AgentConfig{
			BlockingConfig: &traceconfig.BlockingConfig{
				Enabled:  traceconfig.Bool(true),
				DebugLog: traceconfig.Bool(true),
				Modsecurity: &traceconfig.ModsecurityConfig{
					Enabled: traceconfig.Bool(false),
				},
				EvaluateBody: traceconfig.Bool(false),
				RegionBlocking: &traceconfig.RegionBlockingConfig{
					Enabled: traceconfig.Bool(false),
				},
				RemoteConfig: &traceconfig.RemoteConfig{
					Enabled: traceconfig.Bool(false),
				},
			},
			Opa: &traceconfig.Opa{
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
		&traceconfig.AgentConfig{
			BlockingConfig: &traceconfig.BlockingConfig{
				Enabled: traceconfig.Bool(true),
				RemoteConfig: &traceconfig.RemoteConfig{
					Endpoint:          traceconfig.String("agent.traceableai:5441"),
					PollPeriodSeconds: traceconfig.Int32(10),
				},
			},
			Opa: &traceconfig.Opa{
				Endpoint:          traceconfig.String("http://opa:8181"),
				PollPeriodSeconds: traceconfig.Int32(10),
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
