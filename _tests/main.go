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
	cfg.TraceableConfig.DebugLog = config.Bool(true)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Logging is working!")

	f := traceable.NewFilter("", cfg.Tracing.ServiceName.Value, cfg.TraceableConfig, logger)
	if !f.Start() {
		log.Fatal("Failed to initialize traceable filter")
	}

	_, s, ender := goagent.StartSpan(context.Background(), "test")
	defer ender()

	// This run time test does not block because there isn't OPA. For now, we
	// must eyeball that the libtraceable debug output passes the right
	// attributes to the respective calls:
	//   process_request_headers
	//   process_request_body
	s.SetAttribute("http.url", "https://abc.com")
	s.SetAttribute("http.request.header.x-forwarded-for", "83.39.254.157")
	s.SetAttribute("http.request.header.a", "/usr/bin/perl")
	s.SetAttribute("http.request.body", []byte("{\"bad_body\":\"/usr/bin/perl\""))

	_ = f.Evaluate(s)

	if !f.Stop() {
		log.Fatal("Failed to initialize traceable filter")
	}

	fmt.Println("Hello world!")
}
