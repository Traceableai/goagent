package main

import (
	"context"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/filter/traceable"
)

func main() {
	cfg := traceableconfig.Load()
	f := traceable.NewFilter(cfg)

	_, s, ender := goagent.StartSpan(context.Background(), "test")
	defer ender()

	_ = f.EvaluateBody(s, []byte("my_body"))
}
