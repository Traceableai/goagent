//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

// "-Wl,-rpath=\$ORIGIN" ensures we don't need to pass LD_LIBRARY_PATH when running the application.
// See https://stackoverflow.com/a/44214486

/*
#cgo CFLAGS: -I./
#cgo alpine LDFLAGS: -L${SRCDIR}/libs/linux_amd64-alpine -Wl,-rpath=\$ORIGIN -ltraceable -ldl
#cgo !alpine LDFLAGS: -L${SRCDIR}/libs/linux_amd64 -Wl,-rpath=\$ORIGIN -ltraceable -ldl
#include "blocking.h"

#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
	"go.uber.org/zap"
)

const defaultAgentManagerEndpoint = "localhost:5441"
const defaultPollPeriodSec = 30

// NewFilter creates libtraceable based blocking filter
func NewFilter(config *traceableconfig.AgentConfig, logger *zap.Logger) *Filter {
	blockingConfig := config.BlockingConfig
	// disabled if no blocking config or enabled is set to false
	if blockingConfig == nil ||
		blockingConfig.Enabled == nil ||
		blockingConfig.Enabled.Value == false {
		logger.Debug("Traceable filter is disabled by config.")
		return &Filter{logger: logger}
	}

	libTraceableConfig := getLibTraceableConfig(config)
	defer freeLibTraceableConfig(libTraceableConfig)

	var blockingFilter Filter
	ret := C.traceable_new_blocking_engine(libTraceableConfig, &blockingFilter.blockingEngine)
	if ret != C.TRACEABLE_SUCCESS {
		logger.Warn("Failed to initialize traceable filter.")
		return &Filter{logger: logger}
	}

	blockingFilter.logger = logger
	return &blockingFilter
}

type Filter struct {
	blockingEngine C.traceable_blocking_engine
	started        bool
	logger         *zap.Logger
}

var _ filter.Filter = (*Filter)(nil)

// Start() starts the threads to poll config
func (f *Filter) Start() bool {
	if f.blockingEngine != nil {
		ret := C.traceable_start_blocking_engine(f.blockingEngine)
		if ret == C.TRACEABLE_SUCCESS {
			f.started = true
			return true
		}

		f.logger.Warn("Failed to start blocking engine")
	}
	f.logger.Debug("Failed to start with null blocking engine")
	return false
}

func (f *Filter) Stop() bool {
	if f.blockingEngine != nil {
		ret := C.traceable_delete_blocking_engine(f.blockingEngine)
		if ret == C.TRACEABLE_SUCCESS {
			f.started = false
			return true
		}

		f.logger.Warn("Failed to stop blocking engine")
	}
	return false
}

const (
	httpRequestPrefix  = "http.request.header."
	grpcRequestPrefix  = "rpc.request.metadata."
	grpcContentType    = "application/grpc"
	grpcContentTypeLen = 16
)

// isGRPC determines whether a metadata set belongs to http or no
func isGRPC(h map[string][]string) bool {
	contentType, ok := h["Content-Type"]
	if !ok || len(contentType) == 0 {
		return false
	}

	return contentType[0][:grpcContentTypeLen] == grpcContentType
}

// EvaluateURLAndHeaders calls into libtraceable to evaluate if request with URL should be blocked
// or if request with headers should be blocked
func (f *Filter) EvaluateURLAndHeaders(span sdk.Span, url string, headers map[string][]string) bool {
	if !f.started {
		f.logger.Debug("No evaluation of URL or headers as engine isn't started")
		return false
	}

	headerAttributes := map[string]string{
		// evaluate URL together with headers
		"http.url": url,
	}

	prefix := httpRequestPrefix
	if isGRPC(headers) {
		prefix = grpcRequestPrefix
	}

	for k, v := range headers {
		if len(v) == 1 {
			headerAttributes[fmt.Sprintf("%s%s", prefix, strings.ToLower(k))] = v[0]
		} else {
			for i, vv := range v {
				headerAttributes[fmt.Sprintf("%s%s[%d]", prefix, strings.ToLower(k), i)] = vv
			}
		}
	}

	return f.evaluate(span, headerAttributes)
}

// EvaluateBody calls into libtraceable to evaluate if request with body should be blocked
func (f *Filter) EvaluateBody(span sdk.Span, body []byte) bool {
	// no need to call into libtraceable if no body, cgo is expensive.
	if !f.started {
		f.logger.Debug("No evaluation of body as engine isn't started")
		return false
	}

	if len(body) == 0 {
		return false
	}

	return f.evaluate(span, map[string]string{
		"http.request.body": string(body),
	})
}

// evaluate is a common function that calls into libtraceable
// and returns block result attributes to be added to span.
func (f *Filter) evaluate(span sdk.Span, attributes map[string]string) bool {
	inputLibTraceableAttributes := createLibTraceableAttributes(attributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	var blockResult C.traceable_block_result
	ret := C.traceable_block_request(f.blockingEngine, inputLibTraceableAttributes, &blockResult)
	defer C.traceable_delete_block_result_data(blockResult)
	// if call fails just return false
	if ret != C.TRACEABLE_SUCCESS {
		f.logger.Debug("Failed to evaluate attributes")
		return false
	}

	outputAttributes := fromLibTraceableAttributes(blockResult.attributes)
	for k, v := range outputAttributes {
		span.SetAttribute(k, v)
	}

	return blockResult.block != 0
}

// createTraceableAttributes converts map of attributes into C.traceable_attributes
func createLibTraceableAttributes(attributes map[string]string) C.traceable_attributes {
	if len(attributes) == 0 {
		return C.traceable_attributes{
			count:           C.int(len(attributes)),
			attribute_array: (*C.traceable_attribute)(nil),
		}
	}

	var inputAttributes C.traceable_attributes
	inputAttributes.count = C.int(len(attributes))
	inputAttributes.attribute_array = (*C.traceable_attribute)(C.malloc(C.size_t(C.sizeof_traceable_attribute) * C.size_t(len(attributes))))
	i := 0
	for k, v := range attributes {
		inputAttribute := (*C.traceable_attribute)(unsafe.Pointer(uintptr(unsafe.Pointer(inputAttributes.attribute_array)) + uintptr(i*C.sizeof_traceable_attribute)))
		(*inputAttribute).key = C.CString(k)
		(*inputAttribute).value = C.CString(v)
		i++
	}

	return inputAttributes
}

// freeLibTraceableAttributes deletes allocated data in C.traceable_attributes
func freeLibTraceableAttributes(attributes C.traceable_attributes) {
	s := getSliceFromCTraceableAttributes(attributes)
	for _, attribute := range s {
		C.free(unsafe.Pointer(attribute.key))
		C.free(unsafe.Pointer(attribute.value))
	}
	C.free(unsafe.Pointer(attributes.attribute_array))
}

func fromLibTraceableAttributes(attributes C.traceable_attributes) map[string]string {
	s := getSliceFromCTraceableAttributes(attributes)
	m := make(map[string]string)
	for _, attribute := range s {
		m[getGoString(attribute.key)] = getGoString(attribute.value)
	}
	return m
}

func getLibTraceableConfig(config *traceableconfig.AgentConfig) C.traceable_blocking_config {
	blocking, opa := config.BlockingConfig, config.Opa

	// debug log off by default
	opaDebugLog := C.int(0)
	libTraceableLogMode := C.TRACEABLE_LOG_MODE(C.TRACEABLE_LOG_NONE)
	if blocking.DebugLog != nil && blocking.DebugLog.Value {
		libTraceableLogMode = C.TRACEABLE_LOG_MODE(C.TRACEABLE_LOG_STDOUT)
		opaDebugLog = C.int(1)
	}

	logConfig := C.traceable_log_configuration{
		mode: libTraceableLogMode,
	}

	opaConfig := C.traceable_opa_config{
		opa_server_url:      C.CString(opa.Endpoint.Value),
		log_to_console:      C.int(1),
		logging_dir:         C.CString(""),
		logging_file_prefix: C.CString(""),
		cert_file:           C.CString(opa.CertFile.Value),
		debug_log:           opaDebugLog,
		skip_verify:         C.int(0),
		min_delay:           C.int(opa.PollPeriodSeconds.Value),
		max_delay:           C.int(opa.PollPeriodSeconds.Value),
	}

	// modsec on by default
	modsecEnabled := C.int(1)
	if blocking.Modsecurity != nil &&
		blocking.Modsecurity.Enabled != nil &&
		!blocking.Modsecurity.Enabled.Value {
		modsecEnabled = C.int(0)
	}

	modsecurityConfig := C.traceable_modsecurity_config{
		enabled: modsecEnabled,
	}

	// region blocking on by default
	regionBlockingEnabled := C.int(1)
	if blocking.RegionBlocking != nil &&
		blocking.RegionBlocking.Enabled != nil &&
		!blocking.RegionBlocking.Enabled.Value {
		regionBlockingEnabled = C.int(0)
	}

	regionBlockingConfig := C.traceable_rangeblocking_config{
		enabled: regionBlockingEnabled,
	}

	blockingRemoteConfigEnabled := 1
	blockingRemoteConfigEndpoint := defaultAgentManagerEndpoint
	blockingRemoteConfigPollPeriodSec := defaultPollPeriodSec
	blockingRemoteConfigCertFile := ""
	if blocking.GetRemoteConfig() != nil {
		remoteConfig := blocking.GetRemoteConfig()
		if remoteConfig.Enabled != nil && !remoteConfig.GetEnabled().GetValue() {
			blockingRemoteConfigEnabled = 0
			blockingRemoteConfigEndpoint = ""
		} else {
			if remoteConfig.GetEndpoint().GetValue() != "" {
				blockingRemoteConfigEndpoint = remoteConfig.GetEndpoint().GetValue()
			}
			if remoteConfig.GetPollPeriodSeconds() != nil && remoteConfig.GetPollPeriodSeconds().GetValue() != 0 {
				blockingRemoteConfigPollPeriodSec = int(remoteConfig.GetPollPeriodSeconds().GetValue())
			}
			if remoteConfig.GetCertFile() != nil {
				blockingRemoteConfigCertFile = remoteConfig.GetCertFile().GetValue()
			}
		}
	}

	blockingRemoteConfig := C.traceable_remote_config{
		enabled:         C.int(blockingRemoteConfigEnabled),
		remote_endpoint: C.CString(blockingRemoteConfigEndpoint),
		poll_period_sec: C.int(blockingRemoteConfigPollPeriodSec),
		cert_file:       C.CString(blockingRemoteConfigCertFile),
	}

	evaluateBody := C.int(1)
	if blocking.EvaluateBody != nil && !blocking.EvaluateBody.Value {
		evaluateBody = C.int(0)
	}

	skipInternalRequest := C.int(1)
	if blocking.SkipInternalRequest != nil && !blocking.SkipInternalRequest.Value {
		skipInternalRequest = C.int(0)
	}

	return C.traceable_blocking_config{
		log_config:            logConfig,
		opa_config:            opaConfig,
		modsecurity_config:    modsecurityConfig,
		rb_config:             regionBlockingConfig,
		evaluate_body:         evaluateBody,
		remote_config:         blockingRemoteConfig,
		skip_internal_request: skipInternalRequest,
	}
}

func freeLibTraceableConfig(config C.traceable_blocking_config) {
	C.free(unsafe.Pointer(config.opa_config.opa_server_url))
	C.free(unsafe.Pointer(config.opa_config.logging_dir))
	C.free(unsafe.Pointer(config.opa_config.logging_file_prefix))
	C.free(unsafe.Pointer(config.remote_config.remote_endpoint))
}

func getSliceFromCTraceableAttributes(attributes C.traceable_attributes) []C.traceable_attribute {
	if attributes.count == 0 {
		return []C.traceable_attribute{}
	}
	// https://stackoverflow.com/questions/48756732/what-does-1-30c-yourtype-do-exactly-in-cgo
	return (*[1 << 30]C.traceable_attribute)(unsafe.Pointer(attributes.attribute_array))[:attributes.count:attributes.count]
}

func getGoString(cStr *C.char) string {
	return C.GoString(cStr)
}
