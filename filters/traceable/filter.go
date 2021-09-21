//go:build linux
// +build linux

package traceable // import "github.com/Traceableai/goagent/filters/traceable"

// "-Wl,-rpath=\$ORIGIN" ensures we don't need to pass LD_LIBRARY_PATH when running the application.
// See https://stackoverflow.com/a/44214486

/*
#cgo CFLAGS: -I./library
#cgo LDFLAGS: -L${SRCDIR}/../../ -Wl,-rpath=\$ORIGIN -ltraceable -ldl
#include "blocking.h"

#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	traceconfig "github.com/Traceableai/agent-config/gen/go/v1"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
)

const defaultAgentManagerEndpoint = "localhost:5441"
const defaultPollPeriodSec = 30

// NewFilter creates libtraceable based blocking filter
func NewFilter(config *traceconfig.AgentConfig) filter.Filter {
	blockingConfig := config.BlockingConfig
	// disabled if no blocking config or enabled is set to false
	if blockingConfig == nil || blockingConfig.Enabled.Value == false {
		return filter.NoopFilter{}
	}

	libTraceableConfig := getLibTraceableConfig(config)
	defer freeLibTraceableConfig(libTraceableConfig)

	var blockingFilter libTraceableFilter
	ret := C.traceable_new_blocking_engine(libTraceableConfig, &blockingFilter.blockingEngine)
	if ret != C.TRACEABLE_SUCCESS {
		return filter.NoopFilter{}
	}
	return &blockingFilter
}

type libTraceableFilter struct {
	blockingEngine C.traceable_blocking_engine
}

var _ filter.Filter = (*libTraceableFilter)(nil)

// EvaluateURLAndHeaders calls into libtraceable to evaluate if request with URL should be blocked
// or if request with headers should be blocked
func (f *libTraceableFilter) EvaluateURLAndHeaders(span sdk.Span, url string, headers map[string][]string) bool {
	headerAttributes := map[string]string{
		// evaluate URL together with headers
		"http.url": url,
	}
	for k, v := range headers {
		if len(v) == 1 {
			headerAttributes[strings.ToLower(k)] = v[0]
		} else {
			for i, vv := range v {
				headerAttributes[fmt.Sprintf("%s[%d]", strings.ToLower(k), i)] = vv
			}
		}
	}

	return f.evaluate(span, headerAttributes)
}

// EvaluateBody calls into libtraceable to evaluate if request with body should be blocked
func (f *libTraceableFilter) EvaluateBody(span sdk.Span, body []byte) bool {
	// no need to call into libtraceable if no body, cgo is expensive.
	if len(body) == 0 {
		return false
	}

	return f.evaluate(span, map[string]string{
		"http.request.body": string(body),
	})
}

// evaluate is a common function that calls into libtraceable
// and returns block result attributes to be added to span.
func (f *libTraceableFilter) evaluate(span sdk.Span, attributes map[string]string) bool {
	inputLibTraceableAttributes := createLibTraceableAttributes(attributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	var blockResult C.traceable_block_result
	ret := C.traceable_block_request(f.blockingEngine, inputLibTraceableAttributes, &blockResult)
	defer C.traceable_delete_block_result_data(blockResult)
	// if call fails just return false
	if ret != C.TRACEABLE_SUCCESS {
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

func getLibTraceableConfig(config *traceconfig.AgentConfig) C.traceable_blocking_config {
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
		}
	}

	blockingRemoteConfig := C.traceable_remote_config{
		enabled:         C.int(blockingRemoteConfigEnabled),
		remote_endpoint: C.CString(blockingRemoteConfigEndpoint),
		poll_period_sec: C.int(blockingRemoteConfigPollPeriodSec),
	}

	evaluateBody := C.int(1)
	if blocking.EvaluateBody != nil && !blocking.EvaluateBody.Value {
		evaluateBody = C.int(0)
	}

	return C.traceable_blocking_config{
		log_config:         logConfig,
		opa_config:         opaConfig,
		modsecurity_config: modsecurityConfig,
		rb_config:          regionBlockingConfig,
		evaluate_body:      evaluateBody,
		remote_config:      blockingRemoteConfig,
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