package goagent // import "github.com/Traceableai/goagent"

import (
	"context"
	"time"

	"github.com/hypertrace/goagent/instrumentation/hypertrace"
	"github.com/hypertrace/goagent/sdk"
)

//type Span sdk.Span

func SpanFromContext(ctx context.Context) sdk.Span {
	return hypertrace.SpanFromContext(ctx)
}

type Option func(o *sdk.SpanOptions)

type SpanKind string

const (
	SpanKindUndetermined SpanKind = SpanKind(sdk.SpanKindUndetermined)
	SpanKindClient       SpanKind = SpanKind(sdk.SpanKindClient)
	SpanKindServer       SpanKind = SpanKind(sdk.SpanKindServer)
	SpanKindProducer     SpanKind = SpanKind(sdk.SpanKindProducer)
	SpanKindConsumer     SpanKind = SpanKind(sdk.SpanKindConsumer)
)

func WithSpanKind(kind SpanKind) Option {
	return func(o *sdk.SpanOptions) {
		o.Kind = sdk.SpanKind(kind)
	}
}

func WithTimestamp(ts time.Time) Option {
	return func(o *sdk.SpanOptions) {
		o.Timestamp = ts
	}
}

type SpanStarter func(ctx context.Context, name string, opts ...Option) (context.Context, sdk.Span, func())

func translateSpanStarter(s sdk.StartSpan) SpanStarter {
	return func(ctx context.Context, name string, opts ...Option) (context.Context, sdk.Span, func()) {
		o := &sdk.SpanOptions{}
		for _, opt := range opts {
			opt(o)
		}

		return s(ctx, name, o)
	}
}

var StartSpan = translateSpanStarter(hypertrace.StartSpan)
