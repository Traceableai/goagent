package otel // import "github.com/Traceableai/goagent/instrumentation/net/traceablehttp/otel"

import "github.com/hypertrace/goagent/instrumentation/opentelemetry/net/hyperhttp"

var WrapTransport = hyperhttp.WrapTransport
