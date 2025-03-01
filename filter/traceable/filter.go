//go:build linux && traceable_filter

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
	"unsafe"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	goagentconfig "github.com/Traceableai/goagent/config"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64-alpine"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_arm64"
	"github.com/hypertrace/goagent/instrumentation/opentelemetry/identifier"
	"github.com/hypertrace/goagent/sdk"
	"github.com/hypertrace/goagent/sdk/filter"
	filterresult "github.com/hypertrace/goagent/sdk/filter/result"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Filter struct {
	libtraceableHandle C.traceable_libtraceable
	libtraceable       *libtraceableMethods
	started            bool
	logger             *zap.Logger
	responseStatusCode int32
	responseMessage    string
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

var _ filter.Filter = (*Filter)(nil)

// NewFilter creates libtraceable based filter.
// It takes tenant id, service name, agent config and logger as parameters for creating a corresponding filter.
// Library consumers which doesn't have access to tenant id should pass an empty string.
func NewFilter(
	tenantId string,
	serviceName string,
	config *traceableconfig.AgentConfig,
	logger *zap.Logger) *Filter {
	if !config.BlockingConfig.Enabled.Value &&
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
	populateLibtraceableConfig(&libTraceableConfig, tenantId, serviceName, config)
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

	traceableFilter.responseMessage = config.BlockingConfig.ResponseMessage.Value

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

// Evaluate calls into libtraceable to evaluate if request url, body and headers. It is
// EvaluateURLAndHeaders and EvaluateBody combined into one call.
func (f *Filter) Evaluate(span sdk.Span) filterresult.FilterResult {
	if !f.started {
		f.logger.Debug("No evaluation as engine isn't started")
		return filterresult.FilterResult{}
	}

	attributes := map[string]string{}
	span.GetAttributes().Iterate(func(key string, value interface{}) bool {
		attributes[key] = fmt.Sprintf("%v", value)

		// the iterator from ht agent sends values based on this return value
		return true
	})

	inputLibTraceableAttributes := createLibTraceableAttributes(attributes)
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

	return filterresult.FilterResult{
		Block:              processResult.block == 1,
		ResponseStatusCode: f.responseStatusCode,
		ResponseMessage:    f.responseMessage,
		Decorations:        fromLibTraceableDecorations(processResult.decorations),
	}
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

// createLibTraceableStringArray converts a slice of wrapperspb.StringValue to C.traceable_string_array
func createLibTraceableStringArray(values []*wrapperspb.StringValue) C.traceable_string_array {
	if len(values) == 0 {
		return C.traceable_string_array{
			count:  C.int(0),
			values: (**C.char)(nil),
		}
	}
	charPtrSize := unsafe.Sizeof((*C.char)(nil))
	var arr C.traceable_string_array
	arr.count = C.int(len(values))
	arr.values = (**C.char)(C.malloc(C.size_t(charPtrSize) * C.size_t(len(values))))
	i := 0
	for _, value := range values {
		inputPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(arr.values)) + uintptr(i*int(charPtrSize))))
		*inputPtr = C.CString(value.Value)
		i++
	}

	return arr
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

// freeLibTraceableStringArray deletes allocated data in C.traceable_string_array
func freeLibTraceableStringArray(arr C.traceable_string_array) {
	s := getSliceFromCTraceableStringArray(arr)
	for _, val := range s {
		C.free(unsafe.Pointer(val))
	}
	C.free(unsafe.Pointer((**C.char)(arr.values)))
}

func fromLibTraceableAttributes(attributes C.traceable_attributes) map[string]string {
	s := getSliceFromCTraceableAttributes(attributes)
	m := make(map[string]string)
	for _, attribute := range s {
		m[getGoString(attribute.key)] = getGoString(attribute.value)
	}
	return m
}

func fromLibTraceableDecorations(decorations C.traceable_decorations) *filterresult.Decorations {
	ret := &filterresult.Decorations{}
	s := getSliceFromCTraceableHeaderInjections(decorations.request_header_injections)
	for _, header := range s {
		ret.RequestHeaderInjections = append(ret.RequestHeaderInjections, filterresult.KeyValueString{
			Key:   getGoString(header.key),
			Value: getGoString(header.value),
		})
	}
	return ret
}

func populateLibtraceableConfig(
	libtraceableConfig *C.traceable_libtraceable_config,
	tenantId string,
	serviceName string,
	config *traceableconfig.AgentConfig) {
	libtraceableConfig.agent_config.tenant_id = C.CString(tenantId)
	libtraceableConfig.agent_config.environment = C.CString(config.GetEnvironment().GetValue())
	libtraceableConfig.agent_config.service_name = C.CString(serviceName)
	libtraceableConfig.agent_config.agent_token = C.CString(config.GetReporting().GetToken().GetValue())
	libtraceableConfig.agent_config.service_instance_id = C.CString(identifier.ServiceInstanceIDAttr.AsString())
	libtraceableConfig.agent_config.host_name = C.CString(getHostName(config.GetResourceAttributes()))
	// remote_config under blocking is deprecated but need to honor that
	remoteConfigPb := config.BlockingConfig.GetRemoteConfig()
	// if it's not there then look at new location
	if remoteConfigPb.String() == goagentconfig.GetDefaultRemoteConfig().String() {
		remoteConfigPb = config.RemoteConfig
	}

	// disable traces pipeline
	libtraceableConfig.trace_exporter_config.enabled = C.int(0)
	// disable metrics exporter pipeline
	libtraceableConfig.metrics_config.exporter.enabled = C.int(0)

	libtraceableConfig.remote_config.enabled = getCBool(remoteConfigPb.Enabled.Value)
	libtraceableConfig.remote_config.remote_endpoint = C.CString(remoteConfigPb.Endpoint.Value)
	libtraceableConfig.remote_config.poll_period_sec = C.int(remoteConfigPb.PollPeriodSeconds.Value)
	libtraceableConfig.remote_config.cert_file = C.CString(remoteConfigPb.CertFile.Value)
	libtraceableConfig.remote_config.grpc_max_call_recv_msg_size = C.long(remoteConfigPb.GrpcMaxCallRecvMsgSize.Value)
	libtraceableConfig.remote_config.use_secure_connection = getCBool(remoteConfigPb.UseSecureConnection.Value)

	libtraceableConfig.blocking_config.enabled = getCBool(config.BlockingConfig.Enabled.Value)
	libtraceableConfig.blocking_config.modsecurity_config.enabled = getCBool(config.BlockingConfig.Modsecurity.Enabled.Value)
	libtraceableConfig.blocking_config.rb_config.enabled = getCBool(config.BlockingConfig.RegionBlocking.Enabled.Value)
	libtraceableConfig.blocking_config.evaluate_body = getCBool(config.BlockingConfig.EvaluateBody.Value)
	libtraceableConfig.blocking_config.skip_internal_request = getCBool(config.BlockingConfig.SkipInternalRequest.Value)
	libtraceableConfig.blocking_config.max_recursion_depth = C.int(config.BlockingConfig.MaxRecursionDepth.Value)
	libtraceableConfig.blocking_config.eds_config.enabled = getCBool(config.BlockingConfig.EdgeDecisionService.Enabled.Value)
	libtraceableConfig.blocking_config.eds_config.endpoint = C.CString(config.BlockingConfig.EdgeDecisionService.Endpoint.Value)
	libtraceableConfig.blocking_config.eds_config.timeout_ms = C.int(config.BlockingConfig.EdgeDecisionService.TimeoutMs.Value)
	includePathRegexes := createLibTraceableStringArray(config.BlockingConfig.GetEdgeDecisionService().GetIncludePathRegexes())
	libtraceableConfig.blocking_config.eds_config.include_path_regexes = includePathRegexes
	excludePathRegexes := createLibTraceableStringArray(config.BlockingConfig.GetEdgeDecisionService().GetExcludePathRegexes())
	libtraceableConfig.blocking_config.eds_config.exclude_path_regexes = excludePathRegexes

	libtraceableConfig.sampling_config.enabled = getCBool(config.Sampling.Enabled.Value)
	libtraceableConfig.sampling_config.default_rate_limit_config.enabled =
		getCBool(config.Sampling.DefaultRateLimitConfig.Enabled.Value)
	libtraceableConfig.sampling_config.default_rate_limit_config.max_count_global =
		C.int64_t(config.Sampling.DefaultRateLimitConfig.MaxCountGlobal.Value)
	libtraceableConfig.sampling_config.default_rate_limit_config.max_count_per_endpoint =
		C.int64_t(config.Sampling.DefaultRateLimitConfig.MaxCountPerEndpoint.Value)
	libtraceableConfig.sampling_config.default_rate_limit_config.refresh_period =
		C.CString(config.Sampling.DefaultRateLimitConfig.RefreshPeriod.Value)
	libtraceableConfig.sampling_config.default_rate_limit_config.value_expiration_period =
		C.CString(config.Sampling.DefaultRateLimitConfig.ValueExpirationPeriod.Value)
	libtraceableConfig.sampling_config.default_rate_limit_config.span_type =
		getCTraceableSpanType(config.Sampling.DefaultRateLimitConfig.SpanType)

	libtraceableConfig.log_config.mode = getCTraceableLogMode(config.Logging.LogMode)
	libtraceableConfig.log_config.level = getCTraceableLogLevel(config.Logging.LogLevel)
	libtraceableConfig.log_config.file_config.max_files = C.int(config.Logging.LogFile.MaxFiles.Value)
	libtraceableConfig.log_config.file_config.max_file_size = C.int(config.Logging.LogFile.MaxFileSize.Value)
	libtraceableConfig.log_config.file_config.log_file = C.CString(config.Logging.LogFile.FilePath.Value)

	libtraceableConfig.metrics_config.enabled =
		getCBool(config.MetricsConfig.Enabled.Value)
	libtraceableConfig.metrics_config.max_queue_size =
		C.int(config.MetricsConfig.MaxQueueSize.Value)
	libtraceableConfig.metrics_config.endpoint_config.enabled =
		getCBool(config.MetricsConfig.EndpointConfig.Enabled.Value)
	libtraceableConfig.metrics_config.endpoint_config.max_endpoints =
		C.int(config.MetricsConfig.EndpointConfig.MaxEndpoints.Value)
	libtraceableConfig.metrics_config.endpoint_config.logging.enabled =
		getCBool(config.MetricsConfig.EndpointConfig.Logging.Enabled.Value)
	libtraceableConfig.metrics_config.endpoint_config.logging.frequency =
		C.CString(config.MetricsConfig.EndpointConfig.Logging.Frequency.Value)
	libtraceableConfig.metrics_config.logging.enabled =
		getCBool(config.MetricsConfig.Logging.Enabled.Value)
	libtraceableConfig.metrics_config.logging.frequency =
		C.CString(config.MetricsConfig.Logging.Frequency.Value)
	libtraceableConfig.metrics_config.exporter.enabled =
		getCBool(config.MetricsConfig.Exporter.Enabled.Value)
	libtraceableConfig.metrics_config.exporter.server.secure =
		getCBool(config.Reporting.Secure.Value)
	libtraceableConfig.metrics_config.exporter.server.endpoint =
		C.CString(config.Reporting.Endpoint.Value)
	libtraceableConfig.metrics_config.exporter.server.cert_file =
		C.CString(config.Reporting.CertFile.Value)
	libtraceableConfig.metrics_config.exporter.server.export_interval_ms =
		C.int(config.MetricsConfig.Exporter.ExportIntervalMs.Value)
	libtraceableConfig.metrics_config.exporter.server.export_timeout_ms =
		C.int(config.MetricsConfig.Exporter.ExportTimeoutMs.Value)
}

func freeLibTraceableConfig(config C.traceable_libtraceable_config) {
	C.free(unsafe.Pointer(config.remote_config.remote_endpoint))
	C.free(unsafe.Pointer(config.remote_config.cert_file))
	C.free(unsafe.Pointer(config.agent_config.service_name))
	C.free(unsafe.Pointer(config.agent_config.environment))
	C.free(unsafe.Pointer(config.agent_config.agent_token))
	C.free(unsafe.Pointer(config.agent_config.service_instance_id))
	C.free(unsafe.Pointer(config.agent_config.host_name))
	C.free(unsafe.Pointer(config.metrics_config.exporter.server.endpoint))
	C.free(unsafe.Pointer(config.metrics_config.exporter.server.cert_file))
	freeLibTraceableStringArray(config.blocking_config.eds_config.include_path_regexes)
	freeLibTraceableStringArray(config.blocking_config.eds_config.exclude_path_regexes)
}

func getSliceFromCTraceableAttributes(attributes C.traceable_attributes) []C.traceable_attribute {
	return unsafe.Slice(
		(*C.traceable_attribute)(unsafe.Pointer(attributes.attribute_array)),
		int(attributes.count))
}

func getSliceFromCTraceableHeaderInjections(headers C.traceable_header_injections) []C.traceable_key_value_string {
	return unsafe.Slice(
		(*C.traceable_key_value_string)(unsafe.Pointer(headers.header_injections_array)),
		int(headers.count))
}

func getSliceFromCTraceableStringArray(arr C.traceable_string_array) []*C.char {
	return unsafe.Slice((**C.char)(unsafe.Pointer(arr.values)), int(arr.count))
}

func getCTraceableLogMode(logMode traceableconfig.LogMode) C.TRACEABLE_LOG_MODE {
	switch logMode {
	case traceableconfig.LogMode_LOG_MODE_NONE:
		return C.TRACEABLE_LOG_NONE
	case traceableconfig.LogMode_LOG_MODE_STDOUT:
		return C.TRACEABLE_LOG_STDOUT
	case traceableconfig.LogMode_LOG_MODE_FILE:
		return C.TRACEABLE_LOG_FILE
	}
	return C.TRACEABLE_LOG_STDOUT
}

func getCTraceableLogLevel(logLevel traceableconfig.LogLevel) C.TRACEABLE_LOG_LEVEL {
	switch logLevel {
	case traceableconfig.LogLevel_LOG_LEVEL_TRACE:
		return C.TRACEABLE_LOG_LEVEL_TRACE
	case traceableconfig.LogLevel_LOG_LEVEL_DEBUG:
		return C.TRACEABLE_LOG_LEVEL_DEBUG
	case traceableconfig.LogLevel_LOG_LEVEL_INFO:
		return C.TRACEABLE_LOG_LEVEL_INFO
	case traceableconfig.LogLevel_LOG_LEVEL_WARN:
		return C.TRACEABLE_LOG_LEVEL_WARN
	case traceableconfig.LogLevel_LOG_LEVEL_ERROR:
		return C.TRACEABLE_LOG_LEVEL_ERROR
	case traceableconfig.LogLevel_LOG_LEVEL_CRITICAL:
		return C.TRACEABLE_LOG_LEVEL_CRITICAL
	}
	return C.TRACEABLE_LOG_LEVEL_INFO
}

func getCTraceableSpanType(spanType traceableconfig.SpanType) C.TRACEABLE_SPAN_TYPE {
	switch spanType {
	case traceableconfig.SpanType_SPAN_TYPE_NO_SPAN:
		return C.TRACEABLE_NO_SPAN
	case traceableconfig.SpanType_SPAN_TYPE_BARE_SPAN:
		return C.TRACEABLE_BARE_SPAN
	case traceableconfig.SpanType_SPAN_TYPE_FULL_SPAN:
		return C.TRACEABLE_FULL_SPAN
	}
	return C.TRACEABLE_FULL_SPAN
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

func getHostName(resourceAttrs map[string]string) string {
	if val, found := resourceAttrs["host.name"]; found {
		return val
	}
	return ""
}
