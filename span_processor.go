package goagent // import "github.com/Traceableai/goagent"

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

// traceableSpanProcessorWrapper traceableSpanProcessorWrapper drops span
// or converts span to bare span based on traceableai.span_type attribute.
// When span-type enforcement is configured to happen at the collector, the
// span is passed through as-is (still tagged) so the TPA can enforce it.
type traceableSpanProcessorWrapper struct{}

const samplingEnforcementLocationKey = "traceableai.sampling_enforcement_location"

func (*traceableSpanProcessorWrapper) OnStart(parent context.Context, s trace.ReadWriteSpan, delegate trace.SpanProcessor) {
	// nothing to do
	delegate.OnStart(parent, s)
}

func (*traceableSpanProcessorWrapper) OnEnd(s trace.ReadOnlySpan, delegate trace.SpanProcessor) {
	enforceSpanType(spanSamplingEnforcementLocation(s), s, delegate)
}

// spanSamplingEnforcementLocation reads the traceableai.sampling_enforcement_location
// attribute off the span, using the same lookup pattern as traceableai.span_type
// below. libtraceable sets this per-span, sourced from the platform sampling
// config, so enforcement location can differ per rule/service without any
// agent-side config plumbing. Defaults to "AGENT" (enforce locally, today's
// behavior) when the attribute is absent, e.g. older libtraceable versions
// that don't yet emit it.
func spanSamplingEnforcementLocation(s trace.ReadOnlySpan) string {
	for _, attr := range s.Attributes() {
		if attr.Key == samplingEnforcementLocationKey {
			return attr.Value.AsString()
		}
	}
	return "AGENT"
}

// enforceSpanType applies the drop/bare enforcement based on the given
// location. "AGENT" (or anything other than "COLLECTOR"): local drop/bare
// enforcement (today's behavior). "COLLECTOR": let every span through as-is
// (the traceableai.span_type tag rides to the TPA collector, which enforces
// it).
func enforceSpanType(location string, s trace.ReadOnlySpan, delegate trace.SpanProcessor) {
	if location == "COLLECTOR" {
		delegate.OnEnd(&traceableSpanWithoutEnforcementLocation{s})
		return
	}

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
				delegate.OnEnd(&traceableSpanWithoutEnforcementLocation{s})
				return
			case "barespan":
				delegate.OnEnd(&traceableBareSpan{s})
				return
			}
		}
	}

	// no traceableai.span_type found, let the span through
	delegate.OnEnd(&traceableSpanWithoutEnforcementLocation{s})
}

// traceableSpanWithoutEnforcementLocation removes
// traceableai.sampling_enforcement_location from the attributes exposed to
// the delegate. That attribute only exists to tell enforceSpanType what to
// do above; TPA's own collector processors act unconditionally on
// traceableai.span_type alone and never read it, so once consumed here it
// must not leak onto the exported span.
type traceableSpanWithoutEnforcementLocation struct {
	trace.ReadOnlySpan
}

func (s *traceableSpanWithoutEnforcementLocation) Attributes() []attribute.KeyValue {
	src := s.ReadOnlySpan.Attributes()
	attrs := make([]attribute.KeyValue, 0, len(src))
	for _, attr := range src {
		if string(attr.Key) == samplingEnforcementLocationKey {
			continue
		}
		attrs = append(attrs, attr)
	}
	return attrs
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

		// same rationale as traceableSpanWithoutEnforcementLocation above -
		// only consumed by enforceSpanType, must not reach the exporter.
		if key == samplingEnforcementLocationKey {
			continue
		}

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
