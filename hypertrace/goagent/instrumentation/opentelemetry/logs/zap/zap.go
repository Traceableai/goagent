package zap

import (
	"slices"

	agentconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap/zapcore"
)

var _ zapcore.Core = (*core)(nil)

type core struct {
	level      zapcore.Level
	delegate   zapcore.Core
	attributes []zapcore.Field
}

func NewZapCore(name string, cfg *agentconfig.LogsExport, provider log.LoggerProvider) zapcore.Core {
	if !cfg.GetEnabled().GetValue() {
		return zapcore.NewNopCore()
	}

	return &core{
		level:    convertLevel(cfg.GetLevel()),
		delegate: otelzap.NewCore(name, otelzap.WithLoggerProvider(provider)),
	}
}

func (c *core) Enabled(level zapcore.Level) bool {
	return c.level.Enabled(level) && c.delegate.Enabled(level)
}

func (c *core) With(attrs []zapcore.Field) zapcore.Core {
	return &core{
		level:    c.level,
		delegate: c.delegate,
		// TODO we have custom field logic here because of this issue:
		// https://github.com/open-telemetry/opentelemetry-go-contrib/issues/7906
		// once that gets fixed this append should be removed and the delegate should be created with delegate.With
		attributes: append(slices.Clone(c.attributes), attrs...),
	}
}

func (c *core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		ce.AddCore(ent, c)
	}
	return ce
}

func (c *core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// override to avoid this lock on every write
	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/bridges/otelzap/core.go#L235
	// https://github.com/open-telemetry/opentelemetry-go/blob/main/sdk/log/provider.go#L124
	entry.LoggerName = ""
	// TODO this should be removed once the aforementioned issue is fixed
	fields = append(slices.Clone(fields), c.attributes...)
	return c.delegate.Write(entry, fields)
}

func (c *core) Sync() error {
	return c.delegate.Sync()
}

func convertLevel(level agentconfig.LogLevel) zapcore.Level {
	switch level {
	case agentconfig.LogLevel_LOG_LEVEL_DEBUG:
		return zapcore.DebugLevel
	case agentconfig.LogLevel_LOG_LEVEL_WARN:
		return zapcore.WarnLevel
	case agentconfig.LogLevel_LOG_LEVEL_ERROR:
		return zapcore.ErrorLevel
	default:
		// we also redirect trace level to the default because we don't actually want to export
		// trace level logs
		// LogLevel_LOG_LEVEL_UNSPECIFIED
		// LogLevel_LOG_LEVEL_TRACE
		// LogLevel_LOG_LEVEL_INFO
		return zapcore.InfoLevel
	}
}
