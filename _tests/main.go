package main

import (
	"context"

	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent/filters/traceable"
	"github.com/hypertrace/goagent/instrumentation/hypertrace"
	"github.com/hypertrace/goagent/sdk"
)

func main() {
	cfg := traceconfig.Load()
	f := traceable.NewFilter(cfg)

	_, s, ender := hypertrace.StartSpan(context.Background(), "test", &sdk.SpanOptions{})
	defer ender()

	_ = f.EvaluateBody(s, []byte("my_body"))
}
