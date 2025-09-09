package opentelemetry

import (
	"context"
	"log"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/identifier"
	config "github.com/hypertrace/agent-config/gen/go/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

func makeLogsExporterFactory(cfg *config.AgentConfig) func(opts ...ServiceOption) (sdklog.Exporter, error) {
	switch cfg.Reporting.TraceReporterType {
	case config.TraceReporterType_LOGGING:
		// stdout exporter
		return func(_ ...ServiceOption) (sdklog.Exporter, error) {
			return stdoutlog.New()
		}
	case config.TraceReporterType_OTLP_HTTP:
		standardOpts := []otlploghttp.Option{
			otlploghttp.WithEndpoint(cfg.GetReporting().GetEndpoint().GetValue()),
		}

		if !cfg.GetReporting().GetSecure().GetValue() {
			standardOpts = append(standardOpts, otlploghttp.WithInsecure())
		}

		certFile := cfg.GetReporting().GetCertFile().GetValue()
		if len(certFile) > 0 {
			standardOpts = append(standardOpts, otlploghttp.WithTLSClientConfig(createTLSConfig(cfg.GetReporting())))
		}

		return func(opts ...ServiceOption) (sdklog.Exporter, error) {
			serviceOpts := &ServiceOptions{
				headers: make(map[string]string),
			}
			for _, opt := range opts {
				opt(serviceOpts)
			}
			finalOpts := append([]otlploghttp.Option{}, standardOpts...)
			finalOpts = append(finalOpts, otlploghttp.WithHeaders(serviceOpts.headers))

			return otlploghttp.New(context.Background(), finalOpts...)
		}
	default:
		return func(opts ...ServiceOption) (sdklog.Exporter, error) {
			endpoint := cfg.GetReporting().GetMetricEndpoint().GetValue()
			if len(endpoint) == 0 {
				endpoint = cfg.GetReporting().GetEndpoint().GetValue()
			}

			serviceOpts := &ServiceOptions{
				headers: make(map[string]string),
			}
			for _, opt := range opts {
				opt(serviceOpts)
			}

			logsOpts := []otlploggrpc.Option{
				otlploggrpc.WithEndpoint(removeProtocolPrefixForOTLP(endpoint)),
				otlploggrpc.WithHeaders(serviceOpts.headers),
			}

			if !cfg.GetReporting().GetSecure().GetValue() {
				logsOpts = append(logsOpts, otlploggrpc.WithInsecure())
			}

			certFile := cfg.GetReporting().GetCertFile().GetValue()
			if len(certFile) > 0 {
				if tlsCredentials, err := credentials.NewClientTLSFromFile(certFile, ""); err == nil {
					logsOpts = append(logsOpts, otlploggrpc.WithTLSCredentials(tlsCredentials))
				} else {
					log.Printf("error while creating tls credentials from cert path %s: %v", certFile, err)
				}
			}

			if cfg.Reporting.GetEnableGrpcLoadbalancing().GetValue() {
				resolver.SetDefaultScheme("dns")
				logsOpts = append(logsOpts, otlploggrpc.WithServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ]}`))
			}

			if serviceOpts.grpcConn != nil {
				logsOpts = append(logsOpts, otlploggrpc.WithGRPCConn(serviceOpts.grpcConn))
			}

			return otlploggrpc.New(context.Background(), logsOpts...)
		}
	}
}

func initializeLogs(cfg *config.AgentConfig, versionInfoAttrs []attribute.KeyValue, opts ...ServiceOption) func() {

	if !cfg.GetTelemetry().GetLogs().GetEnabled().GetValue() {
		// return no-op function if disabled
		return func() {}
	}

	logsExporterFactory := makeLogsExporterFactory(cfg)
	logsExporter, err := logsExporterFactory(opts...)
	if err != nil {
		log.Fatal(err)
	}

	logsBatchProcessor := sdklog.NewBatchProcessor(logsExporter)

	resourceAttrs := createResources(getResourceAttrsWithServiceName(cfg.ResourceAttributes, cfg.GetServiceName().GetValue()), versionInfoAttrs)
	resourceAttrs = append(resourceAttrs, identifier.ServiceInstanceKeyValue)
	logsResource, err := resource.New(context.Background(), resource.WithAttributes(resourceAttrs...))
	if err != nil {
		log.Fatal(err)
	}
	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithResource(logsResource), sdklog.WithProcessor(logsBatchProcessor))
	global.SetLoggerProvider(loggerProvider)
	return func() {
		err = loggerProvider.Shutdown(context.Background())
		if err != nil {
			log.Printf("an error while calling metrics provider shutdown: %v", err)
		}
		err := logsBatchProcessor.Shutdown(context.Background())
		if err != nil {
			log.Printf("an error while calling metrics reader shutdown: %v", err)
		}
	}
}
