//go:build linux && traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

// "-Wl,-rpath=\$ORIGIN" ensures we don't need to pass LD_LIBRARY_PATH when running the application.
// See https://stackoverflow.com/a/44214486

/*
#cgo CFLAGS: -I./
#cgo LDFLAGS: -Wl,-rpath=\$ORIGIN -ldl
#include "libtraceable.h"
*/
import "C"
import (
	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64-alpine"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_arm64"
)

func getLibTraceableConfig(serviceName string, config *traceableconfig.AgentConfig) C.traceable_libtraceable_config {
	libtraceableConfig := C.traceable_libtraceable_config{}
	populateLibtraceableConfig(&libtraceableConfig, "", serviceName, config)
	return libtraceableConfig
}

func getGoLogMode(mode C.TRACEABLE_LOG_MODE) traceableconfig.LogMode {
	switch mode {
	case C.TRACEABLE_LOG_NONE:
		return traceableconfig.LogMode_LOG_MODE_NONE
	case C.TRACEABLE_LOG_STDOUT:
		return traceableconfig.LogMode_LOG_MODE_STDOUT
	case C.TRACEABLE_LOG_FILE:
		return traceableconfig.LogMode_LOG_MODE_FILE
	}
	return traceableconfig.LogMode_LOG_MODE_UNSPECIFIED
}

func getGoLogLevel(level C.TRACEABLE_LOG_LEVEL) traceableconfig.LogLevel {
	switch level {
	case C.TRACEABLE_LOG_LEVEL_TRACE:
		return traceableconfig.LogLevel_LOG_LEVEL_TRACE
	case C.TRACEABLE_LOG_LEVEL_DEBUG:
		return traceableconfig.LogLevel_LOG_LEVEL_DEBUG
	case C.TRACEABLE_LOG_LEVEL_INFO:
		return traceableconfig.LogLevel_LOG_LEVEL_INFO
	case C.TRACEABLE_LOG_LEVEL_WARN:
		return traceableconfig.LogLevel_LOG_LEVEL_WARN
	case C.TRACEABLE_LOG_LEVEL_ERROR:
		return traceableconfig.LogLevel_LOG_LEVEL_ERROR
	case C.TRACEABLE_LOG_LEVEL_CRITICAL:
		return traceableconfig.LogLevel_LOG_LEVEL_CRITICAL
	}
	return traceableconfig.LogLevel_LOG_LEVEL_UNSPECIFIED
}

func getGoSpanType(spanType C.TRACEABLE_SPAN_TYPE) traceableconfig.SpanType {
	switch spanType {
	case C.TRACEABLE_NO_SPAN:
		return traceableconfig.SpanType_SPAN_TYPE_NO_SPAN
	case C.TRACEABLE_BARE_SPAN:
		return traceableconfig.SpanType_SPAN_TYPE_BARE_SPAN
	case C.TRACEABLE_FULL_SPAN:
		return traceableconfig.SpanType_SPAN_TYPE_FULL_SPAN
	}
	return traceableconfig.SpanType_SPAN_TYPE_UNSPECIFIED
}
