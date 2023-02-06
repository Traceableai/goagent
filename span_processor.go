package goagent // import "github.com/Traceableai/goagent"

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

// traceableSpanProcessorWrapper traceableSpanProcessorWrapper drops span
// or converts span to bare span based on traceableai.span_type attribute
type traceableSpanProcessorWrapper struct{}

func (*traceableSpanProcessorWrapper) OnStart(parent context.Context, s trace.ReadWriteSpan, delegate trace.SpanProcessor) {
	// nothing to do
	delegate.OnStart(parent, s)
}

func (*traceableSpanProcessorWrapper) OnEnd(s trace.ReadOnlySpan, delegate trace.SpanProcessor) {
	for _, attr := range s.Attributes() {
		key := string(attr.Key)

		if key == "traceableai.span_type" {
			value := attr.Value.AsString()
			switch value {
			case "nospan":
				// drop the span by not passing the span to
				// the wrapped span processor
				return
			case "fullspan":
				delegate.OnEnd(s)
				return
			case "barespan":
				delegate.OnEnd(&traceableBareSpan{s})
				return
			}
		}
	}

	// no traceableai.span_type found, let the span through
	delegate.OnEnd(s)
}

var headerPrefixes = []string{
	"http.request.header.",
	"http.response.header.",
	"rpc.request.metadata.",
	"rpc.response.metadata.",
}

var bareSpanHeadersToKeep = []string{
	"x-real-ip",
	"forwarded",
	"x-forwarded-for",
	"x-proxyuser-ip",
	":authority",
	"grpc-status",
	":status",
	":path",
	"content-length",
	"content-type",
	"host",
	"user-agent",
}

var bodyPrefixes = []string{
	"http.request.body",
	"http.response.body",
	"rpc.request.body",
	"rpc.response.body",
}

// traceableBareSpan is a wrapper around a span that removes all request response header and body
// attributes, with some exception
type traceableBareSpan struct {
	trace.ReadOnlySpan
}

func (s *traceableBareSpan) Attributes() []attribute.KeyValue {
	attrs := []attribute.KeyValue{}
	for _, attr := range s.ReadOnlySpan.Attributes() {
		key := string(attr.Key)

		shouldRemove := false

		// check if attribute is body attribute
		for _, prefix := range bodyPrefixes {
			if strings.HasPrefix(key, prefix) {
				shouldRemove = true
				break
			}
		}

		// check if attribute is header attribute
		for _, prefix := range headerPrefixes {
			if strings.HasPrefix(key, prefix) {
				// remove all headers
				shouldRemove = true

				for _, headerToKeep := range bareSpanHeadersToKeep {
					if strings.Contains(key, headerToKeep) {
						shouldRemove = false
						break
					}
				}
				break
			}
		}

		if !shouldRemove {
			attrs = append(attrs, attr)
		}
	}

	return attrs
}
