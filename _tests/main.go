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
	cfg.Blocking.BlockingConfig.DebugLog = config.Bool(true)

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

	// This run time test does not block because there isn't OPA. For now, we
	// must eyeball that the libtraceable debug output passes the right
	// attributes to the respective calls:
	//   process_request_headers
	//   process_request_body
	_ = f.EvaluateURLAndHeaders(s, "http://abc.com", map[string][]string{
		"x-forwarded-for":        []string{"83.39.254.157"}, // arbitrary non local test IP
		"http.request.headers.a": []string{"/usr/bin/perl"},
	})

	_ = f.EvaluateBody(s, []byte("{\"bad_body\":\"/usr/bin/perl\""), map[string][]string{
		"x-forwarded-for": []string{"83.39.254.157"}, // arbitrary non local test IP
	})
	if !f.Stop() {
		log.Fatal("Failed to initialize traceable filter")
	}

	fmt.Println("Hello world!")
}
