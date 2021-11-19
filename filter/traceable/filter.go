//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

// "-Wl,-rpath=\$ORIGIN" ensures we don't need to pass LD_LIBRARY_PATH when running the application.
// See https://stackoverflow.com/a/44214486

// The following C wrappers allow us to fail gracefuly whenever we want to start the app but libtraceable
// hasn't been loaded correctly. See https://stackoverflow.com/a/44042537 for more details.

/*
#cgo CFLAGS: -I./
#cgo LDFLAGS: -Wl,-rpath=\$ORIGIN -ldl
#include "blocking.h"
#include <dlfcn.h>
#include <stdlib.h>

typedef TRACEABLE_RET (*traceable_new_blocking_engine_type)(traceable_blocking_config,traceable_blocking_engine*);

TRACEABLE_RET w_traceable_new_blocking_engine(
    void* f,
    traceable_blocking_config blocking_config,
    traceable_blocking_engine* out_blocking_engine
) {
	return ((traceable_new_blocking_engine_type) f)(blocking_config, out_blocking_engine);
}

typedef TRACEABLE_RET (*traceable_start_blocking_engine_type)(traceable_blocking_engine);

TRACEABLE_RET w_traceable_start_blocking_engine (
	traceable_start_blocking_engine_type f,
	traceable_blocking_engine blocking_engine
) {
	return f(blocking_engine);
}

typedef TRACEABLE_RET (*traceable_delete_blocking_engine_type)(traceable_blocking_engine);

TRACEABLE_RET w_traceable_delete_blocking_engine (
	traceable_delete_blocking_engine_type f,
	traceable_blocking_engine blocking_engine
) {
	return f(blocking_engine);
}

typedef TRACEABLE_RET (*traceable_block_request_type)(
	traceable_blocking_engine,
	traceable_attributes,
	traceable_block_result*
);

TRACEABLE_RET w_traceable_block_request (
	traceable_block_request_type f,
	traceable_blocking_engine blocking_engine,
	traceable_attributes attributes,
	traceable_block_result* out_block_result
) {
	return f(blocking_engine, attributes, out_block_result);
}

typedef TRACEABLE_RET (*traceable_delete_block_result_data_type)(traceable_block_result);

TRACEABLE_RET w_traceable_delete_block_result_data (
	traceable_delete_block_result_data_type f,
	traceable_block_result result
) {
	return f(result);
}
*/
import "C"
import (
	"errors"
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

type Filter struct {
	blockingEngine C.traceable_blocking_engine
	blockingLib    *blocking
	started        bool
	logger         *zap.Logger
}

type blocking struct {
	startEngine           C.traceable_start_blocking_engine_type
	deleteEngine          C.traceable_delete_blocking_engine_type
	blockRequest          C.traceable_block_request_type
	deleteBlockResultData C.traceable_delete_block_result_data_type
}

var _ filter.Filter = (*Filter)(nil)

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

	libPath, err := resolveLibPath()
	if err != nil {
		logger.Warn("Failed to resolve path for libtraceable.so", zap.Error(err))
		return &Filter{logger: logger}
	}

	blockingLib := C.dlopen(C.CString(libPath), C.RTLD_NOW)
	if err := C.dlerror(); err != nil {
		logger.Warn(
			"Traceable filter is disabled because library can't be loaded",
			zap.String("traceableai.goagent.lib_path", libPath),
			zap.Error(errors.New(C.GoString(err))),
		)
		return &Filter{logger: logger}
	}

	libTraceableConfig := getLibTraceableConfig(config)
	defer freeLibTraceableConfig(libTraceableConfig)

	var blockingFilter Filter
	res := C.w_traceable_new_blocking_engine(
		C.dlsym(blockingLib, C.CString("traceable_new_blocking_engine")),
		libTraceableConfig,
		&blockingFilter.blockingEngine,
	)
	if res != C.TRACEABLE_SUCCESS {
		logger.Warn(
			"Traceable filter is disabled because engine can't be created.",
			zap.String("traceableai.goagent.lib_path", libPath),
		)
		return &Filter{logger: logger}
	}

	blockingFilter.logger = logger
	blockingFilter.blockingLib, err = loadBlockingLibMethods(blockingLib)
	if err != nil {
		logger.Warn("Traceable filter is disabled.", zap.Error(err))
		return &Filter{logger: logger}
	}

	logger.Debug(
		"Traceable filter enabled successfuly",
		zap.String("traceableai.goagent.lib_path", libPath),
	)

	return &blockingFilter
}

func loadBlockingLibMethods(blockingLib unsafe.Pointer) (*blocking, error) {
	b := blocking{}

	if startEngine := C.dlsym(blockingLib, C.CString("traceable_start_blocking_engine")); startEngine == nil {
		return nil, errors.New("failed to load traceable_start_blocking_engine")
	} else {
		b.startEngine = C.traceable_start_blocking_engine_type(startEngine)
	}

	if deleteEngine := C.dlsym(blockingLib, C.CString("traceable_delete_blocking_engine")); deleteEngine == nil {
		return nil, errors.New("failed to load traceable_delete_blocking_engine")
	} else {
		b.deleteEngine = C.traceable_delete_blocking_engine_type(deleteEngine)
	}

	if blockRequest := C.dlsym(blockingLib, C.CString("traceable_block_request")); blockRequest == nil {
		return nil, errors.New("failed to load traceable_block_request")
	} else {
		b.blockRequest = C.traceable_block_request_type(blockRequest)
	}

	if deleteBlockResultData := C.dlsym(blockingLib, C.CString("traceable_delete_block_result_data")); deleteBlockResultData == nil {
		return nil, errors.New("failed to load traceable_delete_block_result_data")
	} else {
		b.deleteBlockResultData = C.traceable_delete_block_result_data_type(deleteBlockResultData)
	}

	return &b, nil
}

// Start() starts the threads to poll config
func (f *Filter) Start() bool {
	if f.blockingEngine != nil {
		ret := C.w_traceable_start_blocking_engine(f.blockingLib.startEngine, f.blockingEngine)
		if ret == C.TRACEABLE_SUCCESS {
			f.started = true
			return true
		}

		f.logger.Warn("Failed to start blocking engine")
		return false
	}

	f.logger.Debug("Filter started as NOOP because of null blocking engine")
	return true
}

func (f *Filter) Stop() bool {
	if f.blockingEngine != nil {
		ret := C.w_traceable_delete_blocking_engine(f.blockingLib.deleteEngine, f.blockingEngine)
		if ret == C.TRACEABLE_SUCCESS {
			f.started = false
			return true
		}

		f.logger.Warn("Failed to stop blocking engine")
		return false
	}

	return true
}

const (
	httpRequestHeaderPrefix   = "http.request.header."
	grpcRequestMetadataPrefix = "rpc.request.metadata."
)

func toFQNHeaders(headers map[string][]string, prefix string) map[string]string {
	headerAttributes := map[string]string{}
	for k, v := range headers {
		if len(v) == 1 {
			headerAttributes[fmt.Sprintf("%s%s", prefix, strings.ToLower(k))] = v[0]
		} else {
			for i, vv := range v {
				headerAttributes[fmt.Sprintf("%s%s[%d]", prefix, strings.ToLower(k), i)] = vv
			}
		}
	}
	return headerAttributes
}

// EvaluateURLAndHeaders calls into libtraceable to evaluate if request with URL should be blocked
// or if request with headers should be blocked
func (f *Filter) EvaluateURLAndHeaders(span sdk.Span, url string, headers map[string][]string) bool {
	if !f.started {
		f.logger.Debug("No evaluation of URL or headers as engine isn't started")
		return false
	}

	prefix := httpRequestHeaderPrefix
	if isGRPC(headers) {
		prefix = grpcRequestMetadataPrefix
	}

	headerAttributes := toFQNHeaders(headers, prefix)
	headerAttributes["http.url"] = url

	return f.evaluate(span, headerAttributes)
}

// EvaluateBody calls into libtraceable to evaluate if request with body should be blocked. We need to pass
// the headers as well to still evaluate the body but block in case the headers decide to.
func (f *Filter) EvaluateBody(span sdk.Span, body []byte, headers map[string][]string) bool {
	if !f.started {
		f.logger.Debug("No evaluation of body as engine isn't started")
		return false
	}

	// no need to call into libtraceable if no body, cgo is expensive.
	if len(body) == 0 {
		return false
	}

	headerPrefix := httpRequestHeaderPrefix
	bodyAttributeName := "http.request.body"
	if isGRPC(headers) {
		headerPrefix = grpcRequestMetadataPrefix
		bodyAttributeName = "rpc.request.body"
	}

	headerAttributes := toFQNHeaders(headers, headerPrefix)
	headerAttributes[bodyAttributeName] = string(body)

	return f.evaluate(span, headerAttributes)
}

// evaluate is a common function that calls into libtraceable
// and returns block result attributes to be added to span.
func (f *Filter) evaluate(span sdk.Span, attributes map[string]string) bool {
	inputLibTraceableAttributes := createLibTraceableAttributes(attributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	var blockResult C.traceable_block_result
	ret := C.w_traceable_block_request(f.blockingLib.blockRequest, f.blockingEngine, inputLibTraceableAttributes, &blockResult)
	defer C.w_traceable_delete_block_result_data(f.blockingLib.deleteBlockResultData, blockResult)
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
	opaDebugLog := 0
	libTraceableLogMode := C.TRACEABLE_LOG_MODE(C.TRACEABLE_LOG_NONE)
	if blocking.DebugLog != nil && blocking.DebugLog.Value {
		libTraceableLogMode = C.TRACEABLE_LOG_MODE(C.TRACEABLE_LOG_STDOUT)
		opaDebugLog = 1
	}

	logConfig := C.traceable_log_configuration{
		mode: libTraceableLogMode,
	}

	opaCertFile := ""
	if opa.CertFile != nil {
		opaCertFile = opa.CertFile.Value
	}

	opaConfig := C.traceable_opa_config{
		opa_server_url:      C.CString(opa.Endpoint.Value),
		log_to_console:      C.int(1),
		logging_dir:         C.CString(""),
		logging_file_prefix: C.CString(""),
		cert_file:           C.CString(opaCertFile),
		debug_log:           C.int(opaDebugLog),
		skip_verify:         C.int(0),
		min_delay:           C.int(opa.PollPeriodSeconds.Value),
		max_delay:           C.int(opa.PollPeriodSeconds.Value),
	}

	// modsec on by default
	modsecEnabled := 1
	if blocking.Modsecurity != nil &&
		blocking.Modsecurity.Enabled != nil &&
		!blocking.Modsecurity.Enabled.Value {
		modsecEnabled = 0
	}

	modsecurityConfig := C.traceable_modsecurity_config{
		enabled: C.int(modsecEnabled),
	}

	// region blocking on by default
	regionBlockingEnabled := 1
	if blocking.RegionBlocking != nil &&
		blocking.RegionBlocking.Enabled != nil &&
		!blocking.RegionBlocking.Enabled.Value {
		regionBlockingEnabled = 0
	}

	regionBlockingConfig := C.traceable_rangeblocking_config{
		enabled: C.int(regionBlockingEnabled),
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

	evaluateBody := 1
	if blocking.EvaluateBody != nil && !blocking.EvaluateBody.Value {
		evaluateBody = 0
	}

	skipInternalRequest := 1
	if blocking.SkipInternalRequest != nil && !blocking.SkipInternalRequest.Value {
		skipInternalRequest = 0
	}

	return C.traceable_blocking_config{
		log_config:            logConfig,
		opa_config:            opaConfig,
		modsecurity_config:    modsecurityConfig,
		rb_config:             regionBlockingConfig,
		evaluate_body:         C.int(evaluateBody),
		remote_config:         blockingRemoteConfig,
		skip_internal_request: C.int(skipInternalRequest),
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
