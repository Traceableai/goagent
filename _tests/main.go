package main

import (
	"context"

	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/filters/traceable"
)

func main() {
	cfg := traceconfig.Load()
	f := traceable.NewFilter(cfg)

	_, s, ender := goagent.StartSpan(context.Background(), "test")
	defer ender()

	_ = f.EvaluateBody(s, []byte("my_body"))
}
