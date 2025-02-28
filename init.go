package goagent // import "github.com/Traceableai/goagent"

import (
	"os"

	"go.uber.org/zap"

	"github.com/Traceableai/goagent/config"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	"github.com/Traceableai/goagent/version"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var versionInfoAttributes = []attribute.KeyValue{
	semconv.TelemetrySDKNameKey.String("traceable"),
	semconv.TelemetrySDKVersionKey.String(version.Version),
}

// Init initializes Traceable tracing and returns a shutdown function to flush data immediately
// on a termination signal.
func Init(cfg *config.AgentConfig) func() {
	loggerCloser := logger.InitLogger(os.Getenv("TA_LOG_LEVEL"))
	internalstate.AppendCloser(loggerCloser)
	return InitWithAttributesAndZap(cfg, versionInfoAttributes, logger.GetLogger())
}

func InitWithAttributesAndZap(cfg *config.AgentConfig, attributes []attribute.KeyValue, logger *zap.Logger) func() {
	if cfg.Tracing.Enabled.Value {
		internalstate.InitConfig(cfg)
	} else {
		internalstate.InitConfig(internalconfig.DisabledConfig)
	}

	tracingCloser := opentelemetry.InitWithSpanProcessorWrapperAndZap(cfg.Tracing, &traceableSpanProcessorWrapper{}, attributes, logger)
	internalstate.AppendCloser(tracingCloser)
	return internalstate.CloserFn()
}

func RegisterService(
	key string,
	resourceAttributes map[string]string,
	opts ...opentelemetry.ServiceOption) (SpanStarter, trace.TracerProvider, error) {
	s, tp, err := opentelemetry.RegisterServiceWithSpanProcessorWrapper(key, resourceAttributes, &traceableSpanProcessorWrapper{},
		versionInfoAttributes, opts...)
	if err != nil {
		return nil, tp, err
	}

	return translateSpanStarter(s), tp, nil
}
