package main

import (
	"context"
	"fmt"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/filter/traceable"
	"go.uber.org/zap"
)

func main() {
	cfg := traceableconfig.Load()

	f := traceable.NewFilter(cfg, zap.NewNop())
	f.Start()

	_, s, ender := goagent.StartSpan(context.Background(), "test")
	defer ender()

	_ = f.EvaluateBody(s, []byte("my_body"))
	f.Stop()

	fmt.Println("Hello world!")
}
