package opentelemetry

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/Traceableai/goagent/hypertrace/goagent/config"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func newNoopSpan() trace.Span {
	_, noopSpan := noop.NewTracerProvider().Tracer("noop").Start(context.Background(), "test_name")
	return noopSpan
}

func TestIsNoop(t *testing.T) {
	span := &Span{newNoopSpan()}
	assert.True(t, span.IsNoop())

	shutdown := Init(config.Load())
	defer shutdown()

	_, delegateSpan := otel.Tracer(TracerDomain).Start(context.Background(), "test_span")
	span = &Span{delegateSpan}
	assert.False(t, span.IsNoop())
}

func TestMapSpanKind(t *testing.T) {
	assert.Equal(t, trace.SpanKindClient, mapSpanKind(sdk.SpanKindClient))
	assert.Equal(t, trace.SpanKindServer, mapSpanKind(sdk.SpanKindServer))
	assert.Equal(t, trace.SpanKindProducer, mapSpanKind(sdk.SpanKindProducer))
	assert.Equal(t, trace.SpanKindConsumer, mapSpanKind(sdk.SpanKindConsumer))
}

func TestSetAttributeSuccess(t *testing.T) {
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	s.SetAttribute("test_key_1", true)
	s.SetAttribute("test_key_2", int64(1))
	s.SetAttribute("test_key_3", float64(1.2))
	s.SetAttribute("test_key_4", "abc")
	s.SetAttribute("test_key_4", errors.New("xyz"))
}

func TestSpanHasSameServiceInstanceId(t *testing.T) {
	_, original, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	firstId := original.GetAttributes().GetValue("service.instance.id")
	for i := 0; i < 300; i++ {
		_, anotherSpan, _ := StartSpan(context.Background(), fmt.Sprintf("%s%d", "test_span", i), &sdk.SpanOptions{})
		nextId := anotherSpan.GetAttributes().GetValue("service.instance.id")
		assert.Equal(t, firstId, nextId)
	}
}

func TestGenerateAttribute(t *testing.T) {
	assert.Equal(t, attribute.BOOL, generateAttribute("key", true).Value.Type())
	assert.Equal(t, attribute.BOOLSLICE, generateAttribute("key", []bool{true}).Value.Type())
	assert.Equal(t, attribute.INT64, generateAttribute("key", 1).Value.Type())
	assert.Equal(t, attribute.INT64SLICE, generateAttribute("key", []int{1}).Value.Type())
	assert.Equal(t, attribute.INT64, generateAttribute("key", int64(1)).Value.Type())
	assert.Equal(t, attribute.INT64SLICE, generateAttribute("key", []int64{1}).Value.Type())
	assert.Equal(t, attribute.FLOAT64, generateAttribute("key", 1.23).Value.Type())
	assert.Equal(t, attribute.FLOAT64SLICE, generateAttribute("key", []float64{1.23}).Value.Type())
	assert.Equal(t, attribute.STRING, generateAttribute("key", "val").Value.Type())
	assert.Equal(t, attribute.STRINGSLICE, generateAttribute("key", []string{"val"}).Value.Type())

	attr := generateAttribute("key", errors.New("x"))
	assert.Equal(t, attribute.STRING, attr.Value.Type())
	assert.Equal(t, "x", attr.Value.AsString())
}

func TestAddEvent(t *testing.T) {
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	m := make(map[string]interface{})
	s.AddEvent("test_event_1", time.Now(), m)
	m["k1"] = "v1"
	s.AddEvent("test_event_2", time.Now(), m)
	m["k2"] = 23
	s.AddEvent("test_event_3", time.Now(), m)
	m["k3"] = true
	s.AddEvent("test_event_4", time.Now(), m)
}

func TestGetAttributesNoopSpan(t *testing.T) {
	_, s, _ := NoopStartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	s.SetAttribute("string_key", "string_value")

	// as this is no op span the attributes cannot be retrieved
	attrs := s.GetAttributes()
	assert.Equal(t, nil, attrs.GetValue("string_key"))
}

func TestGetAttributes(t *testing.T) {
	sampler := sdktrace.AlwaysSample()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	s.SetAttribute("string_key", "string_value")
	attrs := s.GetAttributes()
	assert.Equal(t, "string_value", attrs.GetValue("string_key"))
	assert.Equal(t, nil, attrs.GetValue("non_existent"))
}

func TestGetResourceAttributesNoopSpan(t *testing.T) {
	_, s, _ := NoopStartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	// as this is no op span the attributes cannot be retrieved
	attrs := s.GetResourceAttributes()
	assert.Zero(t, attrs.Len())
}

func TestGetResourceAttributes(t *testing.T) {
	sampler := sdktrace.AlwaysSample()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes("test", attribute.String("traceable-key", "traceable-value"))),
	)
	otel.SetTracerProvider(tp)
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	assert.Equal(t, 1, s.GetResourceAttributes().Len())
	assert.Equal(t, "traceable-value", s.GetResourceAttributes().GetValue("traceable-key"))
}

func TestIterate(t *testing.T) {
	sampler := sdktrace.AlwaysSample()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	s.SetAttribute("k1", "v1")
	s.SetAttribute("k2", 200)

	numAttrs := 0
	s.GetAttributes().Iterate(func(key string, value interface{}) bool {
		if key == "k1" {
			assert.Equal(t, "v1", fmt.Sprintf("%v", value))
		} else if key == "k2" {
			assert.Equal(t, "200", fmt.Sprintf("%v", value))
		}
		numAttrs++
		return true
	})
	// service.instance.id is added implicitly in StartSpan. span name and kind are also added as attributes. so 5 attributes will be present.
	assert.Equal(t, 5, numAttrs)
}

func TestLen(t *testing.T) {
	sampler := sdktrace.AlwaysSample()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	s.SetAttribute("k1", "v1")
	s.SetAttribute("k2", 200)

	// service.instance.id is added implicitly in StartSpan. span name and kind are also added as attributes. so 5 attributes will be present.
	assert.Equal(t, 5, s.GetAttributes().Len())
}

func TestGetSpanID(t *testing.T) {
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{})
	spanId := s.GetSpanId()
	assert.NotEqual(t, 0, len(spanId))
}

func TestSpanMetadata(t *testing.T) {
	sampler := sdktrace.AlwaysSample()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes("test", attribute.String("traceable-key", "traceable-value"))),
	)
	otel.SetTracerProvider(tp)
	_, s, _ := StartSpan(context.Background(), "test_span", &sdk.SpanOptions{
		Kind: sdk.SpanKindClient,
	})

	attrs := s.GetAttributes()
	// name, kind and service.instance.id
	assert.Equal(t, 3, attrs.Len())
	assert.Equal(t, "test_span", attrs.GetValue("span.name"))
	// span.kind attribute value is lowercased in startSpan function.
	assert.Equal(t, "client", attrs.GetValue("span.kind"))
}

func TestGenerateAttributeInvalidString(t *testing.T) {
	logger := zaptest.NewLogger(t)
	zap.ReplaceGlobals(logger)
	defer zap.ReplaceGlobals(zap.NewNop())

	var invalidString string
	header := (*reflect.StringHeader)(unsafe.Pointer(&invalidString))
	header.Data = uintptr(0)
	header.Len = 10

	generateAttribute("test", invalidString)
}
