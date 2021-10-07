package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/filter/traceable"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Logging is working!")

	f := traceable.NewFilter(cfg.Blocking, logger)
	if !f.Start() {
		log.Fatal("Failed to initialize traceable filter")
	}

	_, s, ender := goagent.StartSpan(context.Background(), "test")
	defer ender()

	_ = f.EvaluateBody(s, []byte("my_body"))
	if !f.Stop() {
		log.Fatal("Failed to initialize traceable filter")
	}

	fmt.Println("Hello world!")
}
