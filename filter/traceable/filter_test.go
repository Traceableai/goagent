//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable

import (
	"testing"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/stretchr/testify/assert"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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

func TestTraceableConfigDisabled(t *testing.T) {
	f := NewFilter("test-service",
		&traceableconfig.AgentConfig{
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled: traceableconfig.Bool(false),
			},
			ApiDiscovery: &traceableconfig.ApiDiscoveryConfig{
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
			DebugLog: traceableconfig.Bool(true),
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
				RemoteConfig: &traceableconfig.RemoteConfig{
					Enabled:                traceableconfig.Bool(true),
					Endpoint:               traceableconfig.String("localhost:5441"),
					CertFile:               traceableconfig.String(""),
					PollPeriodSeconds:      traceableconfig.Int32(30),
					GrpcMaxCallRecvMsgSize: traceableconfig.Int32(32 * 1024 * 1024),
				},
			},
			Opa: &traceableconfig.Opa{
				Enabled:           traceableconfig.Bool(true),
				Endpoint:          traceableconfig.String("http://opa:8181"),
				PollPeriodSeconds: traceableconfig.Int32(10),
				CertFile:          traceableconfig.String("/conf/tls.crt"),
			},
			RemoteConfig: &traceableconfig.RemoteConfig{
				Enabled:                traceableconfig.Bool(true),
				Endpoint:               traceableconfig.String("localhost:5441"),
				CertFile:               traceableconfig.String("/conf/tls.crt"),
				PollPeriodSeconds:      traceableconfig.Int32(60),
				GrpcMaxCallRecvMsgSize: traceableconfig.Int32(64 * 1024 * 1024),
			},
			ApiDiscovery: &traceableconfig.ApiDiscoveryConfig{
				Enabled: traceableconfig.Bool(true),
			},
			Sampling: &traceableconfig.SamplingConfig{
				Enabled: traceableconfig.Bool(true),
			},
		},
	)

	assert.Equal(t, "http://opa:8181", getGoString(libTraceableConfig.blocking_config.opa_config.opa_server_url))
	assert.Equal(t, 1, int(libTraceableConfig.log_config.mode))
	assert.Equal(t, 1, int(libTraceableConfig.blocking_config.enabled))
	assert.Equal(t, 10, int(libTraceableConfig.blocking_config.opa_config.min_delay))
	assert.Equal(t, 10, int(libTraceableConfig.blocking_config.opa_config.max_delay))
	assert.Equal(t, 1, int(libTraceableConfig.blocking_config.opa_config.log_to_console))
	assert.Equal(t, 1, int(libTraceableConfig.blocking_config.opa_config.debug_log))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.opa_config.skip_verify))
	assert.Equal(t, "/conf/tls.crt", getGoString(libTraceableConfig.blocking_config.opa_config.cert_file))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.modsecurity_config.enabled))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.evaluate_body))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.skip_internal_request))
	assert.Equal(t, 0, int(libTraceableConfig.blocking_config.rb_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "localhost:5441", getGoString(libTraceableConfig.remote_config.remote_endpoint))
	assert.Equal(t, "/conf/tls.crt", getGoString(libTraceableConfig.remote_config.cert_file))
	assert.Equal(t, int(64*1024*1024), int(libTraceableConfig.remote_config.grpc_max_call_recv_msg_size))
	assert.Equal(t, 60, int(libTraceableConfig.remote_config.poll_period_sec))
	assert.Equal(t, 1, int(libTraceableConfig.api_discovery_config.enabled))
	assert.Equal(t, 1, int(libTraceableConfig.sampling_config.enabled))

	// verify for deprecated RemoteConfig and debug log
	libTraceableConfig = getLibTraceableConfig(
		"test-service",
		&traceableconfig.AgentConfig{
			Opa: &traceableconfig.Opa{
				Enabled:           traceableconfig.Bool(true),
				Endpoint:          traceableconfig.String("http://localhost:8181/"),
				PollPeriodSeconds: traceableconfig.Int32(30),
				CertFile:          traceableconfig.String(""),
			},
			DebugLog: traceableconfig.Bool(false),
			BlockingConfig: &traceableconfig.BlockingConfig{
				Enabled:  traceableconfig.Bool(true),
				DebugLog: traceableconfig.Bool(true),
				Modsecurity: &traceableconfig.ModsecurityConfig{
					Enabled: traceableconfig.Bool(true),
				},
				EvaluateBody: traceableconfig.Bool(true),
				RegionBlocking: &traceableconfig.RegionBlockingConfig{
					Enabled: traceableconfig.Bool(true),
				},
				SkipInternalRequest: traceableconfig.Bool(true),
				RemoteConfig: &traceableconfig.RemoteConfig{
					Enabled:                traceableconfig.Bool(true),
					Endpoint:               traceableconfig.String("agent.traceableai:5441"),
					PollPeriodSeconds:      traceableconfig.Int32(10),
					CertFile:               traceableconfig.String(""),
					GrpcMaxCallRecvMsgSize: traceableconfig.Int32(64 * 1024 * 1024),
				},
			},
			RemoteConfig: &traceableconfig.RemoteConfig{
				Enabled:                traceableconfig.Bool(true),
				Endpoint:               traceableconfig.String("localhost:5441"),
				PollPeriodSeconds:      traceableconfig.Int32(30),
				CertFile:               traceableconfig.String(""),
				GrpcMaxCallRecvMsgSize: traceableconfig.Int32(32 * 1024 * 1024),
			},
			ApiDiscovery: &traceableconfig.ApiDiscoveryConfig{
				Enabled: traceableconfig.Bool(false),
			},
			Sampling: &traceableconfig.SamplingConfig{
				Enabled: traceableconfig.Bool(false),
			},
		},
	)

	assert.Equal(t, 1, int(libTraceableConfig.log_config.mode))
	assert.Equal(t, 1, int(libTraceableConfig.blocking_config.opa_config.debug_log))
	assert.Equal(t, 1, int(libTraceableConfig.remote_config.enabled))
	assert.Equal(t, "agent.traceableai:5441", getGoString(libTraceableConfig.remote_config.remote_endpoint))
	assert.Equal(t, 10, int(libTraceableConfig.remote_config.poll_period_sec))
	assert.Equal(t, "", getGoString(libTraceableConfig.remote_config.cert_file))
	assert.Equal(t, int(64*1024*1024), int(libTraceableConfig.remote_config.grpc_max_call_recv_msg_size))
}

func TestToFQNHeaders(t *testing.T) {
	assert.Empty(t, toFQNHeaders(nil, ""))
	assert.Empty(t, toFQNHeaders(map[string][]string{}, ""))

	fqnHeaders := toFQNHeaders(map[string][]string{"Content-Length": {"10"}}, "prefix.")
	assert.Equal(t, "10", fqnHeaders["prefix.content-length"])

	fqnHeaders = toFQNHeaders(map[string][]string{"Content-Type": {"a", "b"}}, "prefix2.")
	assert.Equal(t, "a", fqnHeaders["prefix2.content-type[0]"])
	assert.Equal(t, "b", fqnHeaders["prefix2.content-type[1]"])

	// semconv.NetPeerIPKey is "net.peer.ip" and semconv.HTTPMethodKey is "http.method"
	fqnHeaders = toFQNHeaders(map[string][]string{string(semconv.NetPeerIPKey): {"10.10.10.10"}, string(semconv.HTTPMethodKey): {"GET"}}, "prefix.")
	assert.Equal(t, "10.10.10.10", fqnHeaders[string(semconv.NetPeerIPKey)])
	assert.Equal(t, "GET", fqnHeaders[string(semconv.HTTPMethodKey)])
	fqnHeaders = toFQNHeaders(map[string][]string{
		"Net.Peer.Ip": {"10.20.10.20"},
		"HTTP.method": {"GET"},
	}, "prefix.")
	assert.Equal(t, "10.20.10.20", fqnHeaders[string(semconv.NetPeerIPKey)])
	assert.Equal(t, "GET", fqnHeaders[string(semconv.HTTPMethodKey)])
}
