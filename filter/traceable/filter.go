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

typedef TRACEABLE_RET (*traceable_process_request_using_token_type)(
	traceable_libtraceable,
	traceable_attributes,
	traceable_token_details,
	traceable_process_request_result*
);

TRACEABLE_RET w_traceable_process_request_using_token (
	traceable_process_request_using_token_type f,
	traceable_libtraceable libtraceable,
	traceable_attributes attributes,
	traceable_token_details token_details,
	traceable_process_request_result* out_result
) {
	return f(libtraceable, attributes, token_details, out_result);
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
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	traceableconfig "github.com/Traceableai/agent-config/gen/go/v1"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_amd64-alpine"
	_ "github.com/Traceableai/goagent/filter/traceable/libs/linux_arm64"
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/identifier"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter"
	filterresult "github.com/Traceableai/goagent/hypertrace/goagent/sdk/filter/result"
	"github.com/Traceableai/goagent/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	metricInstrumentationName  = "traceable_filter"
	bufferPoolGaugeName        = "traceable.agent.filter.buffer.size"
	samplingLoggerInitialLimit = 1
	samplingLoggerThereafter   = 500
	invalidSpansCounterName    = "traceable.agent.filter.invalid.spans"
	internalOnlyTenantIdHeader = "traceableai-tenant-id"
	internalOnlyAgentToken     = "traceableai-agent-token"
	spanNameKey                = "span.name"
	spanKindKey                = "span.kind"
)

var (
	once sync.Once
	// setting 10 days as the default timeout to enforce blocking behavior
	timeout    = 10 * 24 * time.Hour
	noopResult = filterresult.FilterResult{
		Block: false,
	}
	meter               metric.Meter
	invalidSpanCounter  metric.Int64Counter
	isFilterInitialized atomic.Bool
)

type Filter struct {
	libtraceableHandle   C.traceable_libtraceable
	libtraceable         *libtraceableMethods
	logger               *zap.Logger
	responseStatusCode   int32
	responseMessage      string
	spanSanitizationMode traceableconfig.SpanSanitizationMode

	// filter thread pool
	poolEnabled     bool
	poolSize        int
	buffer          chan filterRequest
	poolLogger      *zap.Logger
	bufferSizeGauge metric.Int64ObservableGauge

	// shutdown semantics
	shutdownWg sync.WaitGroup
	isStarted  atomic.Bool
}

type libtraceableMethods struct {
	startEngine             C.traceable_start_libtraceable_type
	deleteEngine            C.traceable_delete_libtraceable_type
	processRequest          C.traceable_process_request_using_token_type
	deleteProcessResultData C.traceable_delete_process_request_result_data_type
	initLibtraceableConfig  C.init_libtraceable_config_type
}

type filterRequest struct {
	aa           sdk.AttributeAccessor
	ctx          context.Context
	responseChan chan<- filterresult.FilterResult
}

var _ filter.Filter = (*Filter)(nil)

// NewFilter creates libtraceable based filter.
// It takes agent config and logger as parameters for creating a corresponding filter.
func NewFilter(
	config *traceableconfig.AgentConfig,
	logger *zap.Logger) *Filter {
	if !config.BlockingConfig.Enabled.Value &&
		!config.Sampling.Enabled.Value {
		logger.Debug("Traceable filter is disabled by config.")
		return &Filter{logger: logger}
	}

	if isFilterInitialized.Load() {
		logger.Error("Traceable filter is already initialized, returning empty filter object.")
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
	populateLibtraceableConfig(&libTraceableConfig, config)
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
		logger.Warn(
			"The blocking status code should be of form 4xx.",
			zap.Int32("Invalid code-", config.BlockingConfig.ResponseStatusCode.Value),
		)
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

	// initialize worker pool
	traceableFilter.initializeWorkerPool(config.GetGoagent().GetFilterThreadPool())
	once.Do(func() {
		meter = otel.GetMeterProvider().Meter(
			metricInstrumentationName,
			metric.WithInstrumentationVersion(version.Version))
		invalidSpanCounter, err = meter.Int64Counter(invalidSpansCounterName)
		if err != nil {
			logger.Warn("error initializing counter", zap.String("name", invalidSpansCounterName), zap.Error(err))
			invalidSpanCounter = noop.Int64Counter{}
		}
	})

	traceableFilter.spanSanitizationMode = config.GetGoagent().GetSpanSanitizationMode()
	isFilterInitialized.Store(true)
	return &traceableFilter
}

func (f *Filter) initializeWorkerPool(cfg *traceableconfig.ThreadPool) {
	if !cfg.GetEnabled().GetValue() {
		return
	}

	f.poolEnabled = true
	f.poolSize = int(cfg.GetNumWorkers().GetValue())
	bufferSize := int(cfg.GetBufferSize().GetValue())
	if timeoutMs := cfg.GetTimeoutMs().GetValue(); timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	f.logger.Info(
		"initializing filter worker pool",
		zap.Int("pool size", f.poolSize),
		zap.Int("buffer size", bufferSize),
		zap.String("timeout", timeout.String()),
	)

	f.buffer = make(chan filterRequest, bufferSize)
	f.poolLogger = f.logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSamplerWithOptions(core, time.Second, samplingLoggerInitialLimit, samplingLoggerThereafter)
	}))
	if gauge, err := otel.GetMeterProvider().Meter(
		metricInstrumentationName,
		metric.WithInstrumentationVersion(version.Version)).
		Int64ObservableGauge(
			bufferPoolGaugeName,
			metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
				observer.Observe(int64(len(f.buffer)))
				return nil
			})); err == nil {
		f.bufferSizeGauge = gauge
	}

	// start worker pool
	for range f.poolSize {
		go f.consume()
	}
}

func (f *Filter) consume() {
	for {
		req := <-f.buffer
		req.responseChan <- f.evaluate(req.ctx, req.aa)
		close(req.responseChan)
	}
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

	cStrProcessRequest := C.CString("traceable_process_request_using_token")
	defer C.free(unsafe.Pointer(cStrProcessRequest))
	if processRequest := C.dlsym(libHandle, cStrProcessRequest); processRequest == nil {
		return nil, errors.New("failed to load traceable_process_request")
	} else {
		b.processRequest = C.traceable_process_request_using_token_type(processRequest)
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
			f.isStarted.Store(true)
			return true
		}

		f.logger.Warn("Failed to start libtraceable")
		return false
	}

	f.logger.Debug("Filter started as NOOP because of null libtraceable")
	return true
}

func (f *Filter) Stop() error {
	f.logger.Info("Received shutdown signal for traceable filter")
	// set the init flag to false so that no new requests are accepted
	f.isStarted.Store(false)
	f.shutdownWg.Wait()

	defer isFilterInitialized.Store(false)
	if f.libtraceableHandle == nil {
		return nil
	}

	ret := C.w_traceable_delete_libtraceable(f.libtraceable.deleteEngine, f.libtraceableHandle)
	if ret == C.TRACEABLE_SUCCESS {
		f.logger.Info("Successfully shutdown traceable filter")
		return nil
	}
	return errors.New("failed to shutdown libtraceable")
}

// Evaluate adds an incoming filter request to the common buffer if the filter pool is enabled or directly
// evaluates the request if the pool is disabled. It waits when there are no workers free and the buffer is full.
func (f *Filter) Evaluate(ctx context.Context, aa sdk.AttributeAccessor) filterresult.FilterResult {
	if !f.isStarted.Load() {
		f.logger.Debug("Traceable filter not initialized")
		return noopResult
	}
	f.shutdownWg.Add(1)
	defer f.shutdownWg.Done()

	if !f.poolEnabled {
		return f.evaluate(ctx, aa)
	}

	responseChan := make(chan filterresult.FilterResult, 1)
	req := filterRequest{
		responseChan: responseChan,
		aa:           aa,
		ctx:          ctx,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case f.buffer <- req:
	case <-ctx.Done():
		f.poolLogger.Info("Throttled log: filter buffer is full and request timed out waiting for worker, returning default response")
		close(responseChan)
		return noopResult
	}
	return <-responseChan
}

// evaluate calls into libtraceable to evaluate if request url, body and headers. It is
// EvaluateURLAndHeaders and EvaluateBody combined into one call.
func (f *Filter) evaluate(ctx context.Context, aa sdk.AttributeAccessor) (res filterresult.FilterResult) {
	if f.spanSanitizationMode == traceableconfig.SpanSanitizationMode_SPAN_SANITIZATION_MODE_DROP_SPAN_ON_FAILURE {
		defer func() {
			if r := recover(); r != nil {
				// drop the accessor and return noop result
				aa.SetAttribute("traceableai.span_type", "nospan")
				f.logger.Warn("Recovered from invalid attribute for filter evaluation, dropping span")
				res = noopResult
				invalidSpanCounter.Add(ctx, 1)
			}
		}()
	}

	attributes := make(map[string]string, aa.GetAttributes().Len()+aa.GetResourceAttributes().Len())
	aa.GetAttributes().Iterate(func(key string, value interface{}) bool {
		if value == nil {
			f.logger.Warn("Skipping nil attribute for filter evaluation", zap.String("key", key))
			return true
		}

		attributes[key] = fmt.Sprintf("%v", value)
		// the iterator from ht agent sends values based on this return value
		return true
	})

	aa.GetResourceAttributes().Iterate(func(key string, value interface{}) bool {
		if value == nil {
			f.logger.Warn("Skipping nil attribute for filter evaluation", zap.String("key", key))
			return true
		}

		attributes[key] = fmt.Sprintf("%v", value)
		// the iterator from ht agent sends values based on this return value
		return true
	})

	inputLibTraceableAttributes := createLibTraceableAttributes(attributes)
	defer freeLibTraceableAttributes(inputLibTraceableAttributes)

	inputTokenDetails := createLibTraceableTokenDetails(ctx)
	defer freeLibTraceableTokenDetails(inputTokenDetails)

	var processResult C.traceable_process_request_result
	ret := C.w_traceable_process_request_using_token(
		f.libtraceable.processRequest,
		f.libtraceableHandle,
		inputLibTraceableAttributes,
		inputTokenDetails,
		&processResult,
	)
	defer C.w_traceable_delete_process_request_result_data(f.libtraceable.deleteProcessResultData, processResult)
	// if call fails just return false
	if ret != C.TRACEABLE_SUCCESS {
		f.logger.Debug("Failed to evaluate attributes")
		return filterresult.FilterResult{}
	}

	outputAttributes := fromLibTraceableAttributes(processResult.attributes)
	for k, v := range outputAttributes {
		aa.SetAttribute(k, v)
	}

	statusCode := int32(processResult.decorations.response_details.status_code)
	if statusCode == 0 {
		statusCode = f.responseStatusCode
	}

	responseMessage := getGoString(processResult.decorations.response_details.message)
	if responseMessage == "" {
		responseMessage = f.responseMessage
	}

	return filterresult.FilterResult{
		Block:              processResult.block == 1,
		ResponseStatusCode: statusCode,
		ResponseMessage:    responseMessage,
		Decorations:        fromLibTraceableDecorations(processResult.decorations),
		OutAttributes:      outputAttributes,
	}
}

// createTraceableAttributes converts map of attributes into C.traceable_attributes
func createLibTraceableAttributes(attributes map[string]string) C.traceable_attributes {
	if len(attributes) == 0 {
		return C.traceable_attributes{
			count:           C.int(0),
			attribute_array: (*C.traceable_attribute)(nil),
		}
	}

	var inputAttributes C.traceable_attributes
	inputAttributes.count = C.int(len(attributes))
	inputAttributes.attribute_array = (*C.traceable_attribute)(
		C.malloc(C.size_t(C.sizeof_traceable_attribute) * C.size_t(len(attributes))),
	)
	i := 0
	for k, v := range attributes {
		inputAttribute := (*C.traceable_attribute)(
			unsafe.Pointer(
				uintptr(unsafe.Pointer(inputAttributes.attribute_array)) + uintptr(i*C.sizeof_traceable_attribute),
			),
		)
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

// createLibTraceableTokenDetails converts filter.TokenDetails into C.traceable_token_details
func createLibTraceableTokenDetails(ctx context.Context) C.traceable_token_details {
	var token, tenantId string
	md, found := metadata.FromIncomingContext(ctx)
	if found {
		token = mdGetFirstOrDefault(md, internalOnlyAgentToken, "")
		tenantId = mdGetFirstOrDefault(md, internalOnlyTenantIdHeader, "")
	}

	return C.traceable_token_details{
		token:     C.CString(token),
		tenant_id: C.CString(tenantId),
	}
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

// freeLibTraceableTokenDetails deletes allocated data in C.traceable_token_details
func freeLibTraceableTokenDetails(tokenDetails C.traceable_token_details) {
	C.free(unsafe.Pointer(tokenDetails.token))
	C.free(unsafe.Pointer(tokenDetails.tenant_id))
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
	config *traceableconfig.AgentConfig) {
	libtraceableConfig.agent_config.environment = C.CString(config.GetEnvironment().GetValue())
	libtraceableConfig.agent_config.service_name = C.CString(config.GetServiceName().GetValue())
	libtraceableConfig.agent_config.agent_token = C.CString(config.GetReporting().GetToken().GetValue())
	libtraceableConfig.agent_config.service_instance_id = C.CString(identifier.ServiceInstanceIDAttr.AsString())
	libtraceableConfig.agent_config.resource_attributes = createLibTraceableAttributes(config.GetResourceAttributes())
	libtraceableConfig.agent_config.deployment_name = C.CString(config.GetAgentIdentity().GetDeploymentName().GetValue())

	// disable traces pipeline
	libtraceableConfig.trace_exporter_config.enabled = C.int(0)

	remoteConfigPb := config.RemoteConfig
	libtraceableConfig.remote_config.enabled = getCBool(remoteConfigPb.Enabled.Value)
	libtraceableConfig.remote_config.remote_endpoint = C.CString(remoteConfigPb.Endpoint.Value)
	libtraceableConfig.remote_config.poll_period_sec = C.int(remoteConfigPb.PollPeriodSeconds.Value)
	libtraceableConfig.remote_config.cert_file = C.CString(remoteConfigPb.CertFile.Value)
	libtraceableConfig.remote_config.grpc_max_call_recv_msg_size = C.long(remoteConfigPb.GrpcMaxCallRecvMsgSize.Value)
	libtraceableConfig.remote_config.use_secure_connection = getCBool(remoteConfigPb.UseSecureConnection.Value)

	libtraceableConfig.blocking_config.enabled = getCBool(config.BlockingConfig.Enabled.Value)
	libtraceableConfig.blocking_config.modsecurity_config.enabled = getCBool(
		config.BlockingConfig.Modsecurity.Enabled.Value,
	)
	libtraceableConfig.blocking_config.rb_config.enabled = getCBool(config.BlockingConfig.RegionBlocking.Enabled.Value)
	libtraceableConfig.blocking_config.evaluate_body = getCBool(config.BlockingConfig.EvaluateBody.Value)
	libtraceableConfig.blocking_config.skip_internal_request = getCBool(config.BlockingConfig.SkipInternalRequest.Value)
	libtraceableConfig.blocking_config.skip_client_spans = getCBool(config.BlockingConfig.SkipClientSpans.Value)
	libtraceableConfig.blocking_config.max_recursion_depth = C.int(config.BlockingConfig.MaxRecursionDepth.Value)
	libtraceableConfig.blocking_config.eds_config.enabled = getCBool(
		config.BlockingConfig.EdgeDecisionService.Enabled.Value,
	)
	libtraceableConfig.blocking_config.eds_config.endpoint = C.CString(
		config.BlockingConfig.EdgeDecisionService.Endpoint.Value,
	)
	libtraceableConfig.blocking_config.eds_config.timeout_ms = C.int(
		config.BlockingConfig.EdgeDecisionService.TimeoutMs.Value,
	)
	includePathRegexes := createLibTraceableStringArray(
		config.BlockingConfig.GetEdgeDecisionService().GetIncludePathRegexes(),
	)
	libtraceableConfig.blocking_config.eds_config.include_path_regexes = includePathRegexes
	excludePathRegexes := createLibTraceableStringArray(
		config.BlockingConfig.GetEdgeDecisionService().GetExcludePathRegexes(),
	)
	libtraceableConfig.blocking_config.eds_config.exclude_path_regexes = excludePathRegexes
	libtraceableConfig.blocking_config.evaluate_eds_first = getCBool(config.BlockingConfig.EvaluateEdsFirst.Value)

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
	libtraceableConfig.log_config.exporter_config.enabled =
		getCBool(config.GetTelemetry().GetLogs().GetEnabled().GetValue())
	libtraceableConfig.log_config.exporter_config.level =
		getCTraceableLogLevel(config.GetTelemetry().GetLogs().GetLevel())
	libtraceableConfig.log_config.exporter_config.reporter_type =
		getCTraceableReporterType(config.GetReporting().GetTraceReporterType())

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
	libtraceableConfig.metrics_config.exporter.export_interval_ms =
		C.int(config.MetricsConfig.Exporter.ExportIntervalMs.Value)
	libtraceableConfig.metrics_config.exporter.reporter_type =
		getMetricReporterType(config.Reporting.MetricReporterType, config.Reporting.TraceReporterType)

	libtraceableConfig.parser_config.max_body_size =
		C.uint32_t(uint32(config.GetParserConfig().GetMaxBodySize().GetValue()))
	libtraceableConfig.parser_config.graphql.enabled =
		getCBool(config.GetParserConfig().GetGraphql().GetEnabled().GetValue())

	libtraceableConfig.pipeline_manager_config.pipeline_request_queue_size =
		C.int(config.GetPipelineManager().GetPipelineRequestsQueueInitialSize().GetValue())

	libtraceableConfig.detection_config.enabled = getCBool(config.GetDetectionConfig().GetEnabled().GetValue())
	libtraceableConfig.reporting_config.secure =
		getCBool(config.Reporting.Secure.Value)
	libtraceableConfig.reporting_config.endpoint =
		C.CString(config.GetReporting().GetEndpoint().GetValue())
	libtraceableConfig.reporting_config.cert_file =
		C.CString(config.Reporting.CertFile.Value)
	libtraceableConfig.reporting_config.timeout_ms =
		C.int(config.MetricsConfig.Exporter.ExportTimeoutMs.Value)
}

func freeLibTraceableConfig(config C.traceable_libtraceable_config) {
	C.free(unsafe.Pointer(config.remote_config.remote_endpoint))
	C.free(unsafe.Pointer(config.remote_config.cert_file))
	C.free(unsafe.Pointer(config.agent_config.service_name))
	C.free(unsafe.Pointer(config.agent_config.environment))
	C.free(unsafe.Pointer(config.agent_config.agent_token))
	C.free(unsafe.Pointer(config.agent_config.service_instance_id))
	C.free(unsafe.Pointer(config.agent_config.deployment_name))
	C.free(unsafe.Pointer(config.reporting_config.endpoint))
	C.free(unsafe.Pointer(config.reporting_config.cert_file))
	freeLibTraceableStringArray(config.blocking_config.eds_config.include_path_regexes)
	freeLibTraceableStringArray(config.blocking_config.eds_config.exclude_path_regexes)
	freeLibTraceableAttributes(config.agent_config.resource_attributes)
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

func getMetricReporterType(metricReporterType traceableconfig.MetricReporterType, traceReporterType traceableconfig.TraceReporterType) C.TRACEABLE_REPORTER_TYPE {
	if metricReporterType != traceableconfig.MetricReporterType_METRIC_REPORTER_TYPE_UNSPECIFIED {
		return getCTraceableMetricReporterType(metricReporterType)
	}

	return getCTraceableReporterType(traceReporterType)
}

func getCTraceableReporterType(reporterType traceableconfig.TraceReporterType) C.TRACEABLE_REPORTER_TYPE {
	switch reporterType {
	case traceableconfig.TraceReporterType_LOGGING:
		return C.LOGGING
	case traceableconfig.TraceReporterType_OTLP_HTTP:
		return C.OTLP_HTTP
	}
	return C.OTLP
}

func getCTraceableMetricReporterType(reporterType traceableconfig.MetricReporterType) C.TRACEABLE_REPORTER_TYPE {
	switch reporterType {
	case traceableconfig.MetricReporterType_METRIC_REPORTER_TYPE_LOGGING:
		return C.LOGGING
	}
	return C.OTLP
}

func mdGetFirstOrDefault(md metadata.MD, key string, defaultValue string) string {
	if result := md.Get(key); len(result) > 0 && len(result[0]) > 0 {
		return result[0]
	}

	return defaultValue
}
