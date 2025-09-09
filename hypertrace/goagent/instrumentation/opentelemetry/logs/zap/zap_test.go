package zap

import (
	"context"
	"reflect"
	"testing"

	agentconfig "github.com/hypertrace/agent-config/gen/go/v1"
	"github.com/stretchr/testify/assert"
	logapi "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestNewZapCore(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "export enabled",
			enabled: true,
		},
		{
			name:    "export disabled",
			enabled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &agentconfig.LogsExport{
				Enabled: agentconfig.Bool(tt.enabled),
				Level:   agentconfig.LogLevel_LOG_LEVEL_INFO,
			}

			ret := NewZapCore("test", cfg, nil)
			if tt.enabled {
				obj, ok := ret.(*core)
				assert.True(t, ok)
				assert.NotNil(t, obj)
			} else {
				expected := zapcore.NewNopCore()
				assert.True(t, reflect.DeepEqual(ret, expected))
			}
		})
	}
}

func TestConvertLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    agentconfig.LogLevel
		expected zapcore.Level
	}{
		{
			name:     "debug",
			level:    agentconfig.LogLevel_LOG_LEVEL_DEBUG,
			expected: zapcore.DebugLevel,
		},
		{
			name:     "info",
			level:    agentconfig.LogLevel_LOG_LEVEL_INFO,
			expected: zapcore.InfoLevel,
		},
		{
			name:     "warn",
			level:    agentconfig.LogLevel_LOG_LEVEL_WARN,
			expected: zapcore.WarnLevel,
		},
		{
			name:     "error",
			level:    agentconfig.LogLevel_LOG_LEVEL_ERROR,
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "trace",
			level:    agentconfig.LogLevel_LOG_LEVEL_TRACE,
			expected: zapcore.InfoLevel,
		},
		{
			name:     "unspecified",
			level:    agentconfig.LogLevel_LOG_LEVEL_UNSPECIFIED,
			expected: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertLevel(tt.level))
		})
	}
}

func TestEnabled(t *testing.T) {

	logger := zaptest.NewLogger(t)
	core := &core{
		level:    zapcore.InfoLevel,
		delegate: logger.Core(),
	}
	assert.False(t, core.Enabled(zapcore.DebugLevel))
	assert.True(t, core.Enabled(zapcore.InfoLevel))
	assert.True(t, core.Enabled(zapcore.WarnLevel))
	assert.True(t, core.Enabled(zapcore.ErrorLevel))
}

func TestCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)
	testCore := &core{
		level:    zapcore.InfoLevel,
		delegate: logger.Core(),
	}

	tests := []struct {
		name     string
		level    zapcore.Level
		expected *zapcore.CheckedEntry
	}{
		{
			name:     "debug",
			level:    zapcore.DebugLevel,
			expected: &zapcore.CheckedEntry{},
		},
		{
			name:     "info",
			level:    zapcore.InfoLevel,
			expected: (&zapcore.CheckedEntry{}).AddCore(zapcore.Entry{Level: zapcore.InfoLevel}, testCore),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := testCore.Check(zapcore.Entry{Level: tt.level}, &zapcore.CheckedEntry{})
			assert.True(t, reflect.DeepEqual(tt.expected, ce))
		})
	}
}

func TestWrite(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider, exp := getLoggerProvider()
	testCore := NewZapCore(
		"test-logger",
		&agentconfig.LogsExport{
			Enabled: agentconfig.Bool(true),
			Level:   agentconfig.LogLevel_LOG_LEVEL_INFO,
		},
		provider)

	logger = zap.New(zapcore.NewTee(logger.Core(), testCore))

	logger.Debug("debug message")
	logger.Info("info message", zap.String("foo", "info"), zap.Int("bar", 1))
	logger.Warn("warn message", zap.String("foo", "warn"), zap.Int("bar", 2))
	logger.Error("error message", zap.String("foo", "error"), zap.Int("bar", 3))

	assert.Equal(t, 3, len(exp.records))
	assert.Equal(t, "info message", exp.records[0].Body().String())
	assert.Equal(t, "warn message", exp.records[1].Body().String())
	assert.Equal(t, "error message", exp.records[2].Body().String())

	assertionsCount := 0
	exp.records[0].WalkAttributes(func(kv logapi.KeyValue) bool {
		if kv.Key == "foo" {
			assert.Equal(t, "info", kv.Value.String())
			assertionsCount++
		}
		if kv.Key == "bar" {
			assert.Equal(t, int64(1), kv.Value.AsInt64())
			assertionsCount++
		}
		return true
	})
	exp.records[1].WalkAttributes(func(kv logapi.KeyValue) bool {
		if kv.Key == "foo" {
			assert.Equal(t, "warn", kv.Value.String())
			assertionsCount++
		}
		if kv.Key == "bar" {
			assert.Equal(t, int64(2), kv.Value.AsInt64())
			assertionsCount++
		}
		return true
	})
	exp.records[2].WalkAttributes(func(kv logapi.KeyValue) bool {
		if kv.Key == "foo" {
			assert.Equal(t, "error", kv.Value.String())
			assertionsCount++
		}
		if kv.Key == "bar" {
			assert.Equal(t, int64(3), kv.Value.AsInt64())
			assertionsCount++
		}
		return true
	})
	assert.Equal(t, 6, assertionsCount)
}

func getLoggerProvider() (logapi.LoggerProvider, *exporter) {
	exp := &exporter{}

	processor := log.NewSimpleProcessor(exp)
	return log.NewLoggerProvider(log.WithProcessor(processor)), exp
}

var _ log.Exporter = (*exporter)(nil)

type exporter struct {
	records []log.Record
}

func (e *exporter) Export(_ context.Context, records []log.Record) error {
	e.records = append(e.records, records...)
	return nil
}

func (*exporter) Shutdown(context.Context) error {
	return nil
}

func (*exporter) ForceFlush(context.Context) error {
	return nil
}
