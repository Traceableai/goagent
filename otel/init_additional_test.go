package otel_test

import (
	"context"
	"log"
	"net/http"

	"github.com/Traceableai/goagent/config"
	traceablehttp "github.com/Traceableai/goagent/instrumentation/net/traceablehttp/otel"
	traceableotel "github.com/Traceableai/goagent/otel"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var otherSpanExporter trace.SpanExporter = nil

func ExampleInitAsAdditional() {
	hyperSpanProcessor, shutdown := traceableotel.InitAsAdditional(config.Load())
	defer shutdown()

	ctx := context.Background()
	resources, _ := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("my-server"),
		),
	)

	otherSpanProcessor := sdktrace.NewBatchSpanProcessor(
		traceableotel.RemoveGoAgentAttrs(otherSpanExporter),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(hyperSpanProcessor),
		sdktrace.WithSpanProcessor(otherSpanProcessor),
		sdktrace.WithResource(resources),
	)

	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	r := mux.NewRouter()
	r.Handle("/foo", otelhttp.NewHandler(
		traceablehttp.WrapHandler(
			http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}),
		),
		"/foo",
	))

	log.Fatal(http.ListenAndServe(":8081", r))
}
