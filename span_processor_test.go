package goagent

import (
	"context"
	"testing"

	"github.com/Traceableai/goagent/internal/tracetesting"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

type mockSpanProcessor struct {
	spans []trace.ReadOnlySpan
}

func (sp *mockSpanProcessor) OnStart(parent context.Context, s trace.ReadWriteSpan) {
}

func (sp *mockSpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	sp.spans = append(sp.spans, s)
}

func (sp *mockSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (sp *mockSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

func TestSpanProcessorDropSpan(t *testing.T) {
	mockTraceProvider, _ := tracetesting.InitTracer()
	sp := &mockSpanProcessor{}
	spw := &traceableSpanProcessorWrapper{}

	// drop span, span will not be passed to wrapped processor
	_, span := mockTraceProvider.Start(context.Background(), "drop-span")
	span.SetAttributes(attribute.String("traceableai.span_type", "nospan"))
	span.End()
	spw.OnEnd(span.(trace.ReadOnlySpan), sp)
	assert.Equal(t, 0, len(sp.spans))

	// regular span, span will be passed
	_, span = mockTraceProvider.Start(context.Background(), "regular-span")
	span.End()
	spw.OnEnd(span.(trace.ReadOnlySpan), sp)
	assert.Equal(t, 1, len(sp.spans))

	// some unknown span_type
	_, span = mockTraceProvider.Start(context.Background(), "strange-span")
	span.SetAttributes(attribute.String("traceableai.span_type", "something"))
	span.End()
	spw.OnEnd(span.(trace.ReadOnlySpan), sp)
	assert.Equal(t, 2, len(sp.spans))
}

func TestSpanProcessorBareSpan(t *testing.T) {
	mockTraceProvider, _ := tracetesting.InitTracer()
	sp := &mockSpanProcessor{}
	spw := &traceableSpanProcessorWrapper{}

	_, span := mockTraceProvider.Start(context.Background(), "bare-span")
	span.SetAttributes(
		attribute.String("traceableai.span_type", "barespan"),
		attribute.String("http.url", "http://www.abcd.com"),
		attribute.String("http.request.header.x-real-ip", "a"),
		attribute.String("http.request.header.forwarded", "a"),
		attribute.String("http.request.header.x-forwarded-for", "a"),
		attribute.String("http.request.header.x-proxyuser-ip", "a"),
		attribute.String("http.request.header.:authority", "a"),
		attribute.String("rpc.response.metadata.grpc-status", "1"),
		attribute.String("http.response.header.:status", "200"),
		attribute.String("http.request.header.:path", "a"),
		attribute.String("http.request.header.content-length", "10"),
		attribute.String("http.request.header.content-type", "json"),
		attribute.String("http.response.header.content-length", "10"),
		attribute.String("http.response.header.content-type", "json"),
		attribute.String("http.request.header.host", "a"),
		attribute.String("http.request.header.user-agent", "a"),

		// removed
		attribute.String("http.request.header.something", "a"),
		attribute.String("http.response.header.something", "a"),
		attribute.String("rpc.request.metadata.something", "a"),
		attribute.String("rpc.response.metadata.something", "a"),
		attribute.String("http.request.body", "a"),
		attribute.String("http.response.body", "a"),
		attribute.String("rpc.request.body", "a"),
		attribute.String("rpc.response.body", "a"),
	)
	spw.OnEnd(span.(trace.ReadOnlySpan), sp)
	assert.Equal(t, 1, len(sp.spans))
	// 16 attributes left, the rest are removed
	assert.Equal(t, 16, len(sp.spans[0].Attributes()))
	// check the 16 attributes are what we expect
	attrs := tracetesting.LookupAttributes(sp.spans[0].Attributes())
	assert.True(t, attrs.Has("traceableai.span_type"))
	assert.True(t, attrs.Has("http.url"))
	assert.True(t, attrs.Has("http.request.header.x-real-ip"))
	assert.True(t, attrs.Has("http.request.header.forwarded"))
	assert.True(t, attrs.Has("http.request.header.x-forwarded-for"))
	assert.True(t, attrs.Has("http.request.header.x-proxyuser-ip"))
	assert.True(t, attrs.Has("http.request.header.:authority"))
	assert.True(t, attrs.Has("rpc.response.metadata.grpc-status"))
	assert.True(t, attrs.Has("http.response.header.:status"))
	assert.True(t, attrs.Has("http.request.header.:path"))
	assert.True(t, attrs.Has("http.request.header.content-length"))
	assert.True(t, attrs.Has("http.request.header.content-type"))
	assert.True(t, attrs.Has("http.response.header.content-length"))
	assert.True(t, attrs.Has("http.response.header.content-type"))
	assert.True(t, attrs.Has("http.request.header.host"))
	assert.True(t, attrs.Has("http.request.header.user-agent"))
}

func TestSpanProcessorFullSpan(t *testing.T) {
	mockTraceProvider, _ := tracetesting.InitTracer()
	sp := &mockSpanProcessor{}
	spw := &traceableSpanProcessorWrapper{}

	_, span := mockTraceProvider.Start(context.Background(), "full-span")
	span.SetAttributes(
		attribute.String("traceableai.span_type", "fullspan"),
		attribute.String("http.request.header.something", "a"),
		attribute.String("http.response.header.something", "a"),
		attribute.String("rpc.request.metadata.something", "a"),
		attribute.String("rpc.response.metadata.something", "a"),
		attribute.String("http.request.body", "a"),
		attribute.String("http.response.body", "a"),
		attribute.String("rpc.request.body", "a"),
		attribute.String("rpc.response.body", "a"),
	)
	spw.OnEnd(span.(trace.ReadOnlySpan), sp)
	assert.Equal(t, 1, len(sp.spans))
	assert.Equal(t, 9, len(sp.spans[0].Attributes()))
	attrs := tracetesting.LookupAttributes(sp.spans[0].Attributes())
	assert.True(t, attrs.Has("traceableai.span_type"))
	assert.True(t, attrs.Has("http.request.header.something"))
	assert.True(t, attrs.Has("http.response.header.something"))
	assert.True(t, attrs.Has("rpc.request.metadata.something"))
	assert.True(t, attrs.Has("rpc.response.metadata.something"))
	assert.True(t, attrs.Has("http.request.body"))
	assert.True(t, attrs.Has("http.response.body"))
	assert.True(t, attrs.Has("rpc.request.body"))
	assert.True(t, attrs.Has("rpc.response.body"))
}

func TestSpanProcessorNoSpanType(t *testing.T) {
	mockTraceProvider, _ := tracetesting.InitTracer()
	sp := &mockSpanProcessor{}
	spw := &traceableSpanProcessorWrapper{}

	_, span := mockTraceProvider.Start(context.Background(), "unknown-span")
	span.SetAttributes(
		attribute.String("http.request.header.something", "a"),
		attribute.String("http.response.header.something", "a"),
		attribute.String("rpc.request.metadata.something", "a"),
		attribute.String("rpc.response.metadata.something", "a"),
		attribute.String("http.request.body", "a"),
		attribute.String("http.response.body", "a"),
		attribute.String("rpc.request.body", "a"),
		attribute.String("rpc.response.body", "a"),
	)
	spw.OnEnd(span.(trace.ReadOnlySpan), sp)
	assert.Equal(t, 1, len(sp.spans))
	assert.Equal(t, 8, len(sp.spans[0].Attributes()))
	attrs := tracetesting.LookupAttributes(sp.spans[0].Attributes())
	assert.True(t, attrs.Has("http.request.header.something"))
	assert.True(t, attrs.Has("http.response.header.something"))
	assert.True(t, attrs.Has("rpc.request.metadata.something"))
	assert.True(t, attrs.Has("rpc.response.metadata.something"))
	assert.True(t, attrs.Has("http.request.body"))
	assert.True(t, attrs.Has("http.response.body"))
	assert.True(t, attrs.Has("rpc.request.body"))
	assert.True(t, attrs.Has("rpc.response.body"))
}
