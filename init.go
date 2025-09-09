package goagent // import "github.com/Traceableai/goagent"

import (
	"os"

	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"
	internalconfig "github.com/Traceableai/goagent/internal/config"
	"github.com/Traceableai/goagent/internal/logger"
	internalstate "github.com/Traceableai/goagent/internal/state"
	"github.com/Traceableai/goagent/version"
	htconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func InitWithAttributesAndZap(cfg *config.AgentConfig, attributes []attribute.KeyValue, logger *zap.Logger, opts ...opentelemetry.ServiceOption) func() {
	if cfg.Tracing.Enabled.Value {
		internalstate.InitConfig(cfg)
	} else {
		internalstate.InitConfig(internalconfig.DisabledConfig)
	}

	tracingCloser := opentelemetry.InitWithSpanProcessorWrapperAndZap(cfg.Tracing, &traceableSpanProcessorWrapper{}, attributes, logger, opts...)
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

// NewZapCore returns a new [zapcore.Core] which exports the logs to the configured exporter
func NewZapCore(name string, cfg *htconfig.LogsExport) zapcore.Core {
	return opentelemetry.NewZapCore(name, cfg)
}
