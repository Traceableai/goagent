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
#include "libtraceable.h"
#include <dlfcn.h>
#include <stdlib.h>

typedef TRACEABLE_RET (*traceable_new_libtraceable_type)(traceable_libtraceable_config, traceable_libtraceable*);

TRACEABLE_RET w_traceable_new_libtraceable(
    void* f,
    traceable_libtraceable_config config,
    traceable_libtraceable* out_libtraceable
) {
	return ((traceable_new_libtraceable_type) f)(config, out_libtraceable);
}

typedef TRACEABLE_RET (*traceable_start_libtraceable_type)(traceable_libtraceable);

TRACEABLE_RET w_traceable_start_libtraceable (
	traceable_start_libtraceable_type f,
	traceable_libtraceable libtraceable
) {
	return f(libtraceable);
}

typedef TRACEABLE_RET (*traceable_delete_libtraceable_type)(traceable_libtraceable);

TRACEABLE_RET w_traceable_delete_libtraceable (
	traceable_delete_libtraceable_type f,
	traceable_libtraceable libtraceable
) {
	return f(libtraceable);
}

typedef TRACEABLE_RET (*traceable_process_request_headers_type)(
	traceable_libtraceable,
	traceable_attributes,
	traceable_process_request_result*
);

TRACEABLE_RET w_traceable_process_request_headers (
	traceable_process_request_headers_type f,
	traceable_libtraceable libtraceable,
	traceable_attributes attributes,
	traceable_process_request_result* out_result
) {
	return f(libtraceable, attributes, out_result);
}

typedef TRACEABLE_RET (*traceable_process_request_body_type)(
	traceable_libtraceable,
	traceable_attributes,
	traceable_process_request_result*
);

TRACEABLE_RET w_traceable_process_request_body (
	traceable_process_request_body_type f,
	traceable_libtraceable libtraceable,
	traceable_attributes attributes,
	traceable_process_request_result* out_result
) {
	return f(libtraceable, attributes, out_result);
}

typedef TRACEABLE_RET (*traceable_process_request_type)(
	traceable_libtraceable,
	traceable_attributes,
	traceable_process_request_result*
);

TRACEABLE_RET w_traceable_process_request (
	traceable_process_request_type f,
	traceable_libtraceable libtraceable,
	traceable_attributes attributes,
	traceable_process_request_result* out_result
) {
	return f(libtraceable, attributes, out_result);
}

typedef TRACEABLE_RET (*traceable_delete_process_request_result_data_type)(traceable_process_request_result);

TRACEABLE_RET w_traceable_delete_process_request_result_data (
	traceable_delete_process_request_result_data_type f,
	traceable_process_request_result result
) {
	return f(result);
}

typedef traceable_libtraceable_config (*init_libtraceable_config_type)();

traceable_libtraceable_config w_init_libtraceable_config (
	init_libtraceable_config_type f
) {
	return f();
}
*/
import "C"
import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	goagentconfig "github.com/Traceableai/goagent/config"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64-alpine"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_arm64"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
	filterresult "github.com/hypertrace/goagent/sdk/filter/result"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

const (
	defaultAgentManagerEndpoint = "localhost:5441"
	defaultPollPeriodSec        = 30
	httpUrlKey                  = "http.url"
	httpTargetKey               = "http.target"
)

type Filter struct {
	libtraceableHandle C.traceable_libtraceable
	libtraceable       *libtraceableMethods
	started            bool
	logger             *zap.Logger
	responseStatusCode int32
}

type libtraceableMethods struct {
	startEngine             C.traceable_start_libtraceable_type
	deleteEngine            C.traceable_delete_libtraceable_type
	processRequestHeaders   C.traceable_process_request_headers_type
	processRequestBody      C.traceable_process_request_body_type
	processRequest          C.traceable_process_request_type
	deleteProcessResultData C.traceable_delete_process_request_result_data_type
	initLibtraceableConfig  C.init_libtraceable_config_type
}

var URL_ATTRIBUTES = []string{"http.scheme", "net.host.name", "net.host.port", httpTargetKey}

var _ filter.Filter = (*Filter)(nil)

// NewFilter creates libtraceable based filter
func NewFilter(serviceName string, config *traceableconfig.AgentConfig, logger *zap.Logger) *Filter {
	if !config.BlockingConfig.Enabled.Value &&
		!config.ApiDiscovery.Enabled.Value &&
		!config.Sampling.Enabled.Value {
		logger.Debug("Traceable filter is disabled by config.")
		return &Filter{logger: logger}
	}

	libPath, err := resolveLibPath()
	if err != nil {
		logger.Warn("Failed to resolve path for libtraceable.so", zap.Error(err))
		return &Filter{logger: logger}
	}

	cStrLibPath := C.CString(libPath)
	defer C.free(unsafe.Pointer(cStrLibPath))
	libHandle := C.dlopen(cStrLibPath, C.RTLD_NOW)
	if err := C.dlerror(); err != nil {
		logger.Warn(
			"Traceable filter is disabled because library can't be loaded",
			zap.String("traceableai.goagent.lib_path", libPath),
			zap.Error(errors.New(C.GoString(err))),
		)
		return &Filter{logger: logger}
	}

	cStrInitLibtraceableConfig := C.CString("init_libtraceable_config")
	defer C.free(unsafe.Pointer(cStrInitLibtraceableConfig))
	initLibtraceableConfig := C.dlsym(libHandle, cStrInitLibtraceableConfig)
	if initLibtraceableConfig == nil {
		logger.Warn(
			"Traceable filter is disabled because init_libtraceable_config failed to load.")
		return &Filter{logger: logger}
	}

	libTraceableConfig := C.w_init_libtraceable_config(C.init_libtraceable_config_type(initLibtraceableConfig))
	populateLibtraceableConfig(&libTraceableConfig, serviceName, config)
	defer freeLibTraceableConfig(libTraceableConfig)

	var traceableFilter Filter
	cStrNewTraceableConfig := C.CString("traceable_new_libtraceable")
	defer C.free(unsafe.Pointer(cStrNewTraceableConfig))
	res := C.w_traceable_new_libtraceable(
		C.dlsym(libHandle, cStrNewTraceableConfig),
		libTraceableConfig,
		&traceableFilter.libtraceableHandle,
	)
	if res != C.TRACEABLE_SUCCESS {
		logger.Warn(
			"Traceable filter is disabled because engine can't be created.",
			zap.String("traceableai.goagent.lib_path", libPath),
		)
		return &Filter{logger: logger}
	}

	traceableFilter.logger = logger

	// Check if blocking status code is of type 4xx
	if config.BlockingConfig.ResponseStatusCode.Value/100 != 4 {
		logger.Warn("The blocking status code should be of form 4xx.", zap.Int32("Invalid code-", config.BlockingConfig.ResponseStatusCode.Value))
		traceableFilter.responseStatusCode = 403
	} else {
		traceableFilter.responseStatusCode = config.BlockingConfig.ResponseStatusCode.Value
	}

	traceableFilter.libtraceable, err = loadTraceableConfigMethods(libHandle)
	if err != nil {
		logger.Warn("Traceable filter is disabled.", zap.Error(err))
		return &Filter{logger: logger}
	}

	logger.Debug(
		"Traceable filter enabled successfuly",
		zap.String("traceableai.goagent.lib_path", libPath),
	)

	return &traceableFilter
}

func loadTraceableConfigMethods(libHandle unsafe.Pointer) (*libtraceableMethods, error) {
	b := libtraceableMethods{}

	cStrStartTraceableConfig := C.CString("traceable_start_libtraceable")
	defer C.free(unsafe.Pointer(cStrStartTraceableConfig))
	if startEngine := C.dlsym(libHandle, cStrStartTraceableConfig); startEngine == nil {
		return nil, errors.New("failed to load traceable_start_libtraceable")
	} else {
		b.startEngine = C.traceable_start_libtraceable_type(startEngine)
	}

	cStrDeleteTraceableConfig := C.CString("traceable_delete_libtraceable")
	defer C.free(unsafe.Pointer(cStrDeleteTraceableConfig))
	if deleteEngine := C.dlsym(libHandle, cStrDeleteTraceableConfig); deleteEngine == nil {
		return nil, errors.New("failed to load traceable_delete_libtraceable")
	} else {
		b.deleteEngine = C.traceable_delete_libtraceable_type(deleteEngine)
	}

	cStrProcessRequestHeaders := C.CString("traceable_process_request_headers")
	defer C.free(unsafe.Pointer(cStrProcessRequestHeaders))
	if processRequestHeaders := C.dlsym(libHandle, cStrProcessRequestHeaders); processRequestHeaders == nil {
		return nil, errors.New("failed to load traceable_process_request_headers")
	} else {
		b.processRequestHeaders = C.traceable_process_request_headers_type(processRequestHeaders)
	}

	cStrProcessRequestBody := C.CString("traceable_process_request_body")
	defer C.free(unsafe.Pointer(cStrProcessRequestBody))
	if processRequestBody := C.dlsym(libHandle, cStrProcessRequestBody); processRequestBody == nil {
		return nil, errors.New("failed to load traceable_process_request_body")
	} else {
		b.processRequestBody = C.traceable_process_request_body_type(processRequestBody)
	}

	cStrProcessRequest := C.CString("traceable_process_request")
	defer C.free(unsafe.Pointer(cStrProcessRequest))
	if processRequest := C.dlsym(libHandle, cStrProcessRequest); processRequest == nil {
		return nil, errors.New("failed to load traceable_process_request")
	} else {
		b.processRequest = C.traceable_process_request_type(processRequest)
	}

	cStrDeleteProcessRequestResultData := C.CString("traceable_delete_process_request_result_data")
	defer C.free(unsafe.Pointer(cStrDeleteProcessRequestResultData))
	if deleteProcessResultData := C.dlsym(libHandle, cStrDeleteProcessRequestResultData); deleteProcessResultData == nil {
		return nil, errors.New("failed to load traceable_delete_process_request_result_data")
	} else {
		b.deleteProcessResultData = C.traceable_delete_process_request_result_data_type(deleteProcessResultData)
	}

	return &b, nil
}

// Start() starts the threads to poll config
func (f *Filter) Start() bool {
	if f.libtraceableHandle != nil {
		ret := C.w_traceable_start_libtraceable(f.libtraceable.startEngine, f.libtraceableHandle)
		if ret == C.TRACEABLE_SUCCESS {
			f.started = true
			return true
		}

		f.logger.Warn("Failed to start libtraceable")
		return false
	}

	f.logger.Debug("Filter started as NOOP because of null libtraceable")
	return true
}

func (f *Filter) Stop() bool {
	if f.libtraceableHandle != nil {
		ret := C.w_traceable_delete_libtraceable(f.libtraceable.deleteEngine, f.libtraceableHandle)
		if ret == C.TRACEABLE_SUCCESS {
			f.started = false
			return true
		}

		f.logger.Warn("Failed to delete libtraceable")
		return false
	}

	return true
}

const (
	httpRequestHeaderPrefix   = "http.request.header."
	grpcRequestMetadataPrefix = "rpc.request.metadata."
)

func toFQNHeaders(headers map[string][]string, prefix string, span sdk.Span) map[string]string {
	headerAttributes := map[string]string{}
	for k, v := range headers {
		k = strings.ToLower(k)
		// Do not prepend the prefix to some special attributes
		if k == string(semconv.NetPeerIPKey) || k == string(semconv.HTTPMethodKey) {
			headerAttributes[k] = v[0]
		} else if len(v) == 1 {
			headerAttributes[fmt.Sprintf("%s%s", prefix, k)] = v[0]
		} else {
			for i, vv := range v {
				headerAttributes[fmt.Sprintf("%s%s[%d]", prefix, k, i)] = vv
			}
		}
	}

	// add url attributes so that libtraceable can construct the url if needed
	attributes := span.GetAttributes()
	if url := attributes.GetValue(httpUrlKey); url != nil {
		value := fmt.Sprintf("%s", url)
		if strings.Contains(value, "://") {
			headerAttributes[httpUrlKey] = value
		} else {
			// TODO remove after updating ht goagent
			// special case when the span attribute  http.url set is http.target
			if _, found := headerAttributes[httpTargetKey]; !found {
				headerAttributes[httpTargetKey] = value
			}
			for _, key := range URL_ATTRIBUTES {
				if value := attributes.GetValue(key); value != nil {
					headerAttributes[key] = fmt.Sprintf("%v", value)
				}
			}
		}
	}

	return headerAttributes
}

// EvaluateURLAndHeaders calls into libtraceable to evaluate if request with URL should be blocked
// or if request with headers should be blocked
func (f *Filter) EvaluateURLAndHeaders(span sdk.Span, url string, headers map[string][]string) filterresult.FilterResult {
	if !f.started {
		f.logger.Debug("No evaluation of URL or headers as engine isn't started")
		return filterresult.FilterResult{}
	}

	prefix := httpRequestHeaderPrefix
	if isGRPC(headers) {
		prefix = grpcRequestMetadataPrefix
	}

	headerAttributes := toFQNHeaders(headers, prefix, span)
	headerAttributes["http.url"] = url

	inputLibTraceableAttributes := createLibTraceableAttributes(headerAttributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	var processHeadersResult C.traceable_process_request_result
	ret := C.w_traceable_process_request_headers(f.libtraceable.processRequestHeaders, f.libtraceableHandle, inputLibTraceableAttributes, &processHeadersResult)
	defer C.w_traceable_delete_process_request_result_data(f.libtraceable.deleteProcessResultData, processHeadersResult)
	// if call fails just return false
	if ret != C.TRACEABLE_SUCCESS {
		f.logger.Debug("Failed to evaluate header attributes")
		return filterresult.FilterResult{}
	}

	outputAttributes := fromLibTraceableAttributes(processHeadersResult.attributes)
	for k, v := range outputAttributes {
		span.SetAttribute(k, v)
	}

	if processHeadersResult.block == 0 {
		return filterresult.FilterResult{}
	}
	return filterresult.FilterResult{Block: true, ResponseStatusCode: f.responseStatusCode}
}

// EvaluateBody calls into libtraceable to evaluate if request with body should be blocked. We need to pass
// the headers as well to still evaluate the body but block in case the headers decide to.
func (f *Filter) EvaluateBody(span sdk.Span, body []byte, headers map[string][]string) filterresult.FilterResult {
	if !f.started {
		f.logger.Debug("No evaluation of body as engine isn't started")
		return filterresult.FilterResult{}
	}

	// no need to call into libtraceable if no body, cgo is expensive.
	if len(body) == 0 {
		return filterresult.FilterResult{}
	}

	headerPrefix := httpRequestHeaderPrefix
	bodyAttributeName := "http.request.body"
	if isGRPC(headers) {
		headerPrefix = grpcRequestMetadataPrefix
		bodyAttributeName = "rpc.request.body"
	}

	headerAttributes := toFQNHeaders(headers, headerPrefix, span)
	headerAttributes[bodyAttributeName] = string(body)

	inputLibTraceableAttributes := createLibTraceableAttributes(headerAttributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	var processBodyResult C.traceable_process_request_result
	ret := C.w_traceable_process_request_body(f.libtraceable.processRequestBody, f.libtraceableHandle, inputLibTraceableAttributes, &processBodyResult)
	defer C.w_traceable_delete_process_request_result_data(f.libtraceable.deleteProcessResultData, processBodyResult)
	// if call fails just return false
	if ret != C.TRACEABLE_SUCCESS {
		f.logger.Debug("Failed to evaluate body attributes")
		return filterresult.FilterResult{}
	}

	outputAttributes := fromLibTraceableAttributes(processBodyResult.attributes)
	for k, v := range outputAttributes {
		span.SetAttribute(k, v)
	}

	if processBodyResult.block == 0 {
		return filterresult.FilterResult{}
	}
	return filterresult.FilterResult{Block: true, ResponseStatusCode: f.responseStatusCode}
}

// Evaluate calls into libtraceable to evaluate if request url, body and headers. It is
// EvaluateURLAndHeaders and EvaluateBody combined into one call.
func (f *Filter) Evaluate(span sdk.Span, url string, body []byte, headers map[string][]string) filterresult.FilterResult {
	if !f.started {
		f.logger.Debug("No evaluation as engine isn't started")
		return filterresult.FilterResult{}
	}

	headerPrefix := httpRequestHeaderPrefix
	bodyAttributeName := "http.request.body"
	if isGRPC(headers) {
		headerPrefix = grpcRequestMetadataPrefix
		bodyAttributeName = "rpc.request.body"
	}

	headerAttributes := toFQNHeaders(headers, headerPrefix, span)
	headerAttributes["http.url"] = url
	if len(body) > 0 {
		headerAttributes[bodyAttributeName] = string(body)
	}

	inputLibTraceableAttributes := createLibTraceableAttributes(headerAttributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	var processResult C.traceable_process_request_result
	ret := C.w_traceable_process_request(f.libtraceable.processRequest, f.libtraceableHandle, inputLibTraceableAttributes, &processResult)
	defer C.w_traceable_delete_process_request_result_data(f.libtraceable.deleteProcessResultData, processResult)
	// if call fails just return false
	if ret != C.TRACEABLE_SUCCESS {
		f.logger.Debug("Failed to evaluate attributes")
		return filterresult.FilterResult{}
	}

	outputAttributes := fromLibTraceableAttributes(processResult.attributes)
	for k, v := range outputAttributes {
		span.SetAttribute(k, v)
	}

	if processResult.block == 0 {
		return filterresult.FilterResult{}
	}
	return filterresult.FilterResult{Block: true, ResponseStatusCode: f.responseStatusCode}
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

// For unit test only
func getLibTraceableConfig(serviceName string, config *traceableconfig.AgentConfig) C.traceable_libtraceable_config {
	libtraceableConfig := C.traceable_libtraceable_config{}
	populateLibtraceableConfig(&libtraceableConfig, serviceName, config)
	return libtraceableConfig
}

func populateLibtraceableConfig(libtraceableConfig *C.traceable_libtraceable_config, serviceName string, config *traceableconfig.AgentConfig) {
	debugLog := config.DebugLog.Value ||
		// check debug_log in deprecated location too
		config.BlockingConfig.DebugLog.Value

	// debug log off by default
	libTraceableLogMode := C.TRACEABLE_LOG_MODE(C.TRACEABLE_LOG_NONE)
	if debugLog {
		libTraceableLogMode = C.TRACEABLE_LOG_MODE(C.TRACEABLE_LOG_STDOUT)
	}
	libtraceableConfig.log_config.mode = libTraceableLogMode

	// remote_config under blocking is deprecated
	// but need to honor that
	remoteConfigPb := config.BlockingConfig.GetRemoteConfig()
	// if it's not there then look at new location
	if remoteConfigPb.String() == goagentconfig.GetDefaultRemoteConfig().String() {
		remoteConfigPb = config.RemoteConfig
	}

	libtraceableConfig.remote_config.enabled = getCBool(remoteConfigPb.Enabled.Value)
	libtraceableConfig.remote_config.remote_endpoint = C.CString(remoteConfigPb.Endpoint.Value)
	libtraceableConfig.remote_config.poll_period_sec = C.int(remoteConfigPb.PollPeriodSeconds.Value)
	libtraceableConfig.remote_config.cert_file = C.CString(remoteConfigPb.CertFile.Value)
	libtraceableConfig.remote_config.grpc_max_call_recv_msg_size = C.long(remoteConfigPb.GrpcMaxCallRecvMsgSize.Value)
	libtraceableConfig.remote_config.use_secure_connection = getCBool(remoteConfigPb.UseSecureConnection.Value)

	libtraceableConfig.blocking_config.opa_config.enabled = getCBool(config.Opa.Enabled.Value)
	libtraceableConfig.blocking_config.opa_config.opa_server_url = C.CString(config.Opa.Endpoint.Value)
	libtraceableConfig.blocking_config.opa_config.log_to_console = C.int(1)
	libtraceableConfig.blocking_config.opa_config.cert_file = C.CString(config.Opa.CertFile.Value)
	libtraceableConfig.blocking_config.opa_config.debug_log = getCBool(debugLog)
	libtraceableConfig.blocking_config.opa_config.min_delay = C.int(config.Opa.PollPeriodSeconds.Value)
	libtraceableConfig.blocking_config.opa_config.max_delay = C.int(config.Opa.PollPeriodSeconds.Value)
	libtraceableConfig.blocking_config.opa_config.use_secure_connection = getCBool(config.Opa.UseSecureConnection.Value)

	libtraceableConfig.blocking_config.enabled = getCBool(config.BlockingConfig.Enabled.Value)
	libtraceableConfig.blocking_config.modsecurity_config.enabled = getCBool(config.BlockingConfig.Modsecurity.Enabled.Value)
	libtraceableConfig.blocking_config.rb_config.enabled = getCBool(config.BlockingConfig.RegionBlocking.Enabled.Value)
	libtraceableConfig.blocking_config.evaluate_body = getCBool(config.BlockingConfig.EvaluateBody.Value)
	libtraceableConfig.blocking_config.skip_internal_request = getCBool(config.BlockingConfig.SkipInternalRequest.Value)
	libtraceableConfig.blocking_config.max_recursion_depth = C.int(config.BlockingConfig.MaxRecursionDepth.Value)

	libtraceableConfig.agent_config.service_name = C.CString(serviceName)

	libtraceableConfig.api_discovery_config.enabled = getCBool(config.ApiDiscovery.Enabled.Value)

	libtraceableConfig.sampling_config.enabled = getCBool(config.Sampling.Enabled.Value)
}

func freeLibTraceableConfig(config C.traceable_libtraceable_config) {
	C.free(unsafe.Pointer(config.remote_config.remote_endpoint))
	C.free(unsafe.Pointer(config.remote_config.cert_file))
	C.free(unsafe.Pointer(config.blocking_config.opa_config.opa_server_url))
	C.free(unsafe.Pointer(config.blocking_config.opa_config.cert_file))
	C.free(unsafe.Pointer(config.agent_config.service_name))
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

func getCBool(b bool) C.int {
	if b {
		return C.int(1)
	}

	return C.int(0)
}
