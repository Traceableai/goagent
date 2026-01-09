package opentelemetry // import "github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry"

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/opentelemetry/identifier"
	"github.com/Traceableai/goagent/hypertrace/goagent/sdk"
	"github.com/Traceableai/goagent/hypertrace/goagent/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

const (
	spanNameKey = "span.name"
	spanKindKey = "span.kind"
)

var _ sdk.AttributeList = (*AttributeList)(nil)

type AttributeList struct {
	attrs []attribute.KeyValue
}

func (l *AttributeList) GetValue(key string) interface{} {
	for _, attr := range l.attrs {
		if string(attr.Key) == key {
			return attr.Value.AsInterface()
		}
	}

	return nil
}

func (l *AttributeList) Iterate(yield func(key string, value interface{}) bool) {
	for _, attr := range l.attrs {
		if !yield(string(attr.Key), attr.Value.AsInterface()) {
			return
		}
	}
}

func (l *AttributeList) Len() int {
	return len(l.attrs)
}

var _ sdk.Span = (*Span)(nil)

type Span struct {
	trace.Span
}

func generateAttribute(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case bool:
		return attribute.Bool(key, v)
	case []bool:
		if isBadSlice(v) {
			zap.L().Warn("Trying to set invalid bool slice, dropping the attribute", zap.String("key", key))
			return attribute.String(key, "invalid-bool-slice")
		}
		return attribute.BoolSlice(key, v)
	case int:
		return attribute.Int(key, v)
	case []int:
		if isBadSlice(v) {
			zap.L().Warn("Trying to set invalid int slice, dropping the attribute", zap.String("key", key))
			return attribute.String(key, "invalid-int-slice")
		}
		return attribute.IntSlice(key, v)
	case int64:
		return attribute.Int64(key, v)
	case []int64:
		if isBadSlice(v) {
			zap.L().Warn("Trying to set invalid int64 slice, dropping the attribute", zap.String("key", key))
			return attribute.String(key, "invalid-int64-slice")
		}
		return attribute.Int64Slice(key, v)
	case float64:
		return attribute.Float64(key, v)
	case []float64:
		if isBadSlice(v) {
			zap.L().Warn("Trying to set invalid float64 slice, dropping the attribute", zap.String("key", key))
			return attribute.String(key, "invalid-float64-slice")
		}
		return attribute.Float64Slice(key, v)
	case string:
		if isBadString(v) {
			zap.L().Warn("Trying to set invalid string, dropping the attribute", zap.String("key", key))
			return attribute.String(key, "invalid-string")
		}
		return attribute.String(key, v)
	case []string:
		if isBadSlice(v) {
			zap.L().Warn("Trying to set invalid string slice, dropping the attribute", zap.String("key", key))
			return attribute.String(key, "invalid-string-slice")
		}
		return attribute.StringSlice(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

func (s *Span) GetAttributes() sdk.AttributeList {
	if s.IsNoop() {
		return &AttributeList{
			attrs: nil,
		}
	}
	readableSpan := s.Span.(sdktrace.ReadOnlySpan)
	return &AttributeList{
		attrs: readableSpan.Attributes(),
	}
}

func (s *Span) GetResourceAttributes() sdk.AttributeList {
	if s.IsNoop() {
		return &AttributeList{
			attrs: nil,
		}
	}
	readableSpan := s.Span.(sdktrace.ReadOnlySpan)
	return &AttributeList{
		attrs: readableSpan.Resource().Attributes(),
	}
}

func (s *Span) SetAttribute(key string, value interface{}) {
	s.Span.SetAttributes(generateAttribute(key, value))
}

func (s *Span) SetError(err error) {
	s.Span.RecordError(err)
}

func (s *Span) SetStatus(code sdk.Code, description string) {
	s.Span.SetStatus(codes.Code(code), description)
}

func (s *Span) IsNoop() bool {
	return !s.Span.IsRecording()
}

func (s *Span) AddEvent(name string, ts time.Time, attributes map[string]interface{}) {
	otAttributes := []attribute.KeyValue{}
	for k, v := range attributes {
		otAttributes = append(otAttributes, generateAttribute(k, v))
	}
	s.Span.AddEvent(name, trace.WithTimestamp(ts), trace.WithAttributes(otAttributes...))
}

func (s *Span) GetSpanId() string {
	return s.Span.SpanContext().SpanID().String()
}

func (s *Span) GetName() string {
	if roSpan, ok := s.Span.(sdktrace.ReadOnlySpan); ok {
		return roSpan.Name()
	}
	return ""
}

func (s *Span) GetKind() string {
	if roSpan, ok := s.Span.(sdktrace.ReadOnlySpan); ok {
		return roSpan.SpanKind().String()
	}
	return ""
}

func SpanFromContext(ctx context.Context) sdk.Span {
	span := &Span{trace.SpanFromContext(ctx)}
	span.SetAttributes(identifier.ServiceInstanceKeyValue)
	return span
}

type getTracerProvider func() trace.TracerProvider

func NoopTracerProvider() trace.TracerProvider {
	// This returns noop.TracerProvider which though it implements trace.TracerProvider, go compiler complains about its type.
	return noop.NewTracerProvider()
}

func startSpan(provider getTracerProvider) sdk.StartSpan {
	return func(ctx context.Context, name string, opts *sdk.SpanOptions) (context.Context, sdk.Span, func()) {
		startOpts := make([]trace.SpanStartOption, 0)

		if opts != nil {
			startOpts = append(startOpts, trace.WithSpanKind(mapSpanKind(opts.Kind)))
			if opts.Timestamp.IsZero() {
				startOpts = append(startOpts, trace.WithTimestamp(time.Now()))
			} else {
				startOpts = append(startOpts, trace.WithTimestamp(opts.Timestamp))
			}
		} else {
			opts = &sdk.SpanOptions{
				Timestamp: time.Now(),
			}
		}
		startOpts = append(
			startOpts,
			trace.WithAttributes(
				identifier.ServiceInstanceKeyValue,
				attribute.String(spanNameKey, name),
				attribute.String(spanKindKey, strings.ToLower(string(opts.Kind)))))

		ctx, span := provider().
			Tracer(TracerDomain, trace.WithInstrumentationVersion(version.Version)).
			Start(ctx, name, startOpts...)
		return ctx, &Span{span}, func() { span.End() }
	}
}

var StartSpan = startSpan(otel.GetTracerProvider)
var NoopStartSpan = startSpan(NoopTracerProvider)

func mapSpanKind(kind sdk.SpanKind) trace.SpanKind {
	switch kind {
	case sdk.SpanKindClient:
		return trace.SpanKindClient
	case sdk.SpanKindServer:
		return trace.SpanKindServer
	case sdk.SpanKindProducer:
		return trace.SpanKindProducer
	case sdk.SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindUnspecified
	}
}

func isBadString(s string) bool {
	if len(s) == 0 {
		return false
	}
	header := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return header.Data == 0 && header.Len > 0
}

func isBadSlice[T any](s []T) bool {
	if len(s) == 0 {
		return false
	}
	header := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	return header.Data == 0 && header.Len > 0
}
