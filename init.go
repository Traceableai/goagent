package goagent // import "github.com/Traceableai/goagent"

import (
	"log"
	"os"
	"strings"

	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	"github.com/hypertrace/goagent/instrumentation/hypertrace"
	"go.uber.org/zap"
)

// Init initializes Traceable tracing and returns a shutdown function to flush data immediately
// on a termination signal.
func Init(cfg *config.AgentConfig) func() {
	loggerCloser := initLogger(os.Getenv("TA_LOG_LEVEL"))
	internalstate.AppendCloser(loggerCloser)

	internalstate.InitConfig(cfg.Blocking)

	tracingCloser := hypertrace.Init(cfg.Tracing)
	internalstate.AppendCloser(tracingCloser)
	return internalstate.CloserFn()
}

func initLogger(logLevel string) func() {
	var lvl = zap.ErrorLevel
	switch strings.ToLower(logLevel) {
	case "debug":
		lvl = zap.DebugLevel
	case "info":
		lvl = zap.InfoLevel
	case "warn":
		lvl = zap.WarnLevel
	}

	l, err := zap.Config{
		Level:            zap.NewAtomicLevelAt(lvl),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		log.Printf("Failed to init logger: %v", err)
		return func() {}
	}

	logger.InitLogger(l)

	return func() { l.Sync() }
}

func RegisterService(
	serviceName string,
	resourceAttributes map[string]string,
) (SpanStarter, error) {
	s, err := hypertrace.RegisterService(serviceName, resourceAttributes)
	if err != nil {
		return nil, err
	}

	return translateSpanStarter(s), nil
}
